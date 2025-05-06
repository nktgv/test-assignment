package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"http-load-balancer/api"
	"http-load-balancer/balancer"
	"http-load-balancer/configs"
	"http-load-balancer/healthcheck"
	"http-load-balancer/lib/logger/sl"
	"http-load-balancer/lib/strategy"
	"http-load-balancer/limiter"
	"http-load-balancer/repository"
	"http-load-balancer/storage/postgres"
)

func main() {
	cfg := configs.MustLoad()

	log := configs.ConfigureLogger(cfg.Env)

	log.Info("load balancer starting", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	pgStorage, err := postgres.New(
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.DB,
	)
	if err != nil {
		log.Error("failed to init postgres", sl.Err(err))
		os.Exit(1)
	}
	defer pgStorage.Close()
	log.Info("postgres connection established", slog.Int("addr", cfg.Postgres.Port))

	backendRepo := repository.NewBackendRepository(pgStorage.DB)
	userRepo := repository.NewUserRepository(pgStorage.DB)

	backends, err := backendRepo.GetAll()
	if err != nil {
		log.Error("failed to get all backends", sl.Err(err))
		os.Exit(1)
	}
	if len(backends) == 0 {
		for i := range len(cfg.Backends) {
			cfg.Backends[i].IsAlive = true
			backend, err := backendRepo.Add(&cfg.Backends[i])
			log.Debug("backend added", slog.Any("backend", backend))
			if err != nil {
				log.Error("failed to add backend", sl.Err(err))
				os.Exit(1)
			}
			cfg.Backends[i].ID = backend.ID
			cfg.Backends[i].CreatedAt = backend.CreatedAt
			cfg.Backends[i].UpdatedAt = backend.UpdatedAt
		}
	}

	var balancerStrategy strategy.Strategy
	switch cfg.Strategy {
	case "round-robin":
		balancerStrategy = strategy.NewRoundRobin()
	case "least_connections":
		backends, err := backendRepo.GetAll()
		if err != nil {
			log.Error("failed to get all backends", sl.Err(err))
		}
		balancerStrategy, err = strategy.NewLeastConnections(backends)
		if err != nil {
			log.Error("failed to get least connections", sl.Err(err))
		}
	case "random":
		balancerStrategy = strategy.NewRandom()
	default:
		log.Info("Unknown strategy:", slog.String("strategy", cfg.Strategy))
		os.Exit(1)
	}
	log.Info("balancer strategy", slog.String("strategy", cfg.Strategy))

	limiter := limiter.NewTokenBucket(userRepo, cfg.User.DefaultCapacity, cfg.User.DefaultRPS)

	healthchecker := healthcheck.NewHealthChecker(backendRepo, cfg.HealthCheckTimeout)

	balancer := balancer.NewBalancer(
		balancerStrategy,
		backendRepo,
		healthchecker,
		limiter,
		log,
	)

	clientHandler := api.NewClientHandler(userRepo)
	mux := http.NewServeMux()
	mux.Handle("/", balancer)
	mux.HandleFunc("POST /clients", clientHandler.CreateClient)
	mux.HandleFunc("DELETE /clients/{client_id}", clientHandler.DeleteClient)
	mux.HandleFunc("PATCH /clients/{client_id}", clientHandler.UpdateClientParams)

	server := &http.Server{
		Addr:           cfg.Addr + ":" + strconv.Itoa(cfg.Port),
		Handler:        mux,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
		IdleTimeout:    180 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	var serverErrChan = make(chan error)
	log.Info("Server started on port", slog.Int("port", cfg.Port))
	go func() {
		balancer.StartHealthChecks()
		if err := server.ListenAndServe(); err != nil || !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	select {
	case err := <-serverErrChan:
		log.Error("Server error", sl.Err(err))
		os.Exit(1)
	case <-time.After(1 * time.Second):
		log.Info("Server is running")
	}

	<-done
	log.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HealthCheckTimeout)
	defer cancel()

	balancer.StopHealthChecks()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", sl.Err(err))
	}

	log.Info("server stopped")
}
