package balancer

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"http-load-balancer/healthcheck"
	"http-load-balancer/lib/logger/sl"
	"http-load-balancer/lib/strategy"
	"http-load-balancer/limiter"
	"http-load-balancer/models"
	"http-load-balancer/repository"
)

type Balancer struct {
	strategy      strategy.Strategy
	backendRepo   repository.BackendRepository
	healthChecker *healthcheck.HealthChecker
	limiter       *limiter.TokenBucket
	log           *slog.Logger
}

func NewBalancer(
	strategy strategy.Strategy,
	backendRepo repository.BackendRepository,
	healthChecker *healthcheck.HealthChecker,
	limiter *limiter.TokenBucket,
	log *slog.Logger,
) *Balancer {
	return &Balancer{
		strategy,
		backendRepo,
		healthChecker,
		limiter,
		log,
	}
}

func (b *Balancer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	b.log.Debug("incoming request",
		slog.String("method", req.Method),
		slog.String("path", req.URL.Path))
	var userID uint64
	if req.Method == http.MethodPost {
		if req.Body == nil || req.Body == http.NoBody {
			b.log.Error("empty request body")
			http.Error(w, "Request body required", http.StatusBadRequest)
			return
		}

		var tmpUser models.User
		if err := json.NewDecoder(req.Body).Decode(&tmpUser); err != nil {
			b.log.Error("failed to decode request body", sl.Err(err))
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()
		userID = tmpUser.ID
		b.log.Debug("requested userID", slog.Uint64("userID", userID))

		allowed, err := b.limiter.Allow(userID)
		if err != nil {
			b.handleLimiterError(w, err)
			return
		}
		if !allowed {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
	}

	backends, err := b.backendRepo.GetActive()
	if err != nil {
		b.log.Error("failed to get active backends", sl.Err(err))
		if errors.Is(err, repository.ErrNoActiveBackends) {
			b.log.Error("active backends not found", sl.Err(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	b.log.Info("active backends", slog.Any("backends", backends))

	backend, err := b.strategy.NextBackend(backends)
	if err != nil {
		if errors.Is(err, strategy.ErrNoAliveBackends) {
			b.log.Error("active backends not found", sl.Err(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		b.log.Error("failed to select backend", sl.Err(err))
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

	b.proxyRequest(w, req, &backend)
}

func (b *Balancer) StartHealthChecks() {
	b.healthChecker.Start()
}

func (b *Balancer) StopHealthChecks() {
	b.healthChecker.Stop()
}

func (b *Balancer) handleLimiterError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrUserNotFound):
		http.Error(w, "User not found", http.StatusNotFound)
	case errors.Is(err, limiter.ErrRateLimitExceeded):
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
	default:
		b.log.Error("limiter error", sl.Err(err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func (b *Balancer) proxyRequest(w http.ResponseWriter, req *http.Request, backend *models.Backend) {
	target, err := url.Parse("http://" + backend.URL)
	if err != nil {
		b.log.Error("invalid backend URL",
			sl.Err(err),
			slog.String("url", backend.URL))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		b.log.Error("proxy error",
			sl.Err(err),
			slog.String("backend", backend.URL))

		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
	}

	b.log.Debug("proxying request",
		slog.String("method", req.Method),
		slog.String("to", backend.URL))

	req.Header.Set("X-Forwarded-For", req.RemoteAddr)
	req.Header.Set("X-Forwarded-Host", req.Host)

	proxy.ServeHTTP(w, req)
}
