package balancer

import (
	"log/slog"
	"net/http"

	"http-load-balancer/healthcheck"
	"http-load-balancer/lib/strategy"
	"http-load-balancer/limiter"
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

}
