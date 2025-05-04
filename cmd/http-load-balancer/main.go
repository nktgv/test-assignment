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

	limiter := limiter.NewTokenBucket(userRepo, cfg.User.DefaultCapacity, cfg.User.DefaultRPS)

	healthchecker := healthcheck.NewHealthChecker(backendRepo, cfg.HealthCheckTimeout)

	balancer := balancer.NewBalancer(
		balancerStrategy,
		backendRepo,
		healthchecker,
		limiter,
		log,
	)

	mux := http.NewServeMux()
	mux.Handle("/", balancer)
	// mux.HandleFunc("/clients", clientHandler.CreateClient)
	server := &http.Server{
		Addr:    "127.0.0.1:" + strconv.Itoa(cfg.Port),
		Handler: mux,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// balancer.StartHealthChecks()
		log.Info("Server started on port %s", slog.Int("port", cfg.Port))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("Server error: %v", sl.Err(err))
		}
	}()

	<-done
	log.Info("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HealthCheckTimeout)
	defer cancel()

	// balancer.StopHealthChecks()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error: %v", sl.Err(err))
	}

	log.Info("server stopped")
}
