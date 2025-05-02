package main

import (
	"http-load-balancer/configs"
	"http-load-balancer/lib/logger/sl"
	"http-load-balancer/storage/postgres"
	"log/slog"
	"net/http"
	"os"
	"strconv"
)

func main() {
	cfg := configs.MustLoad()

	log := configs.ConfigureLogger(cfg.Env)

	log.Info("load balancer starting", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	_, err := postgres.New(
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

	log.Info("postgres connection established", slog.Int("addr", cfg.Postgres.Port))

	srv := &http.Server{
		Addr: "localhost:" + strconv.Itoa(cfg.Postgres.Port),
	}

	log.Info("starting server", slog.String("address", "localhost:"+strconv.Itoa(cfg.Postgres.Port)))

	err = srv.ListenAndServe()
	if err != nil {
		log.Error("failed to start server")
	}

	log.Error("server stopped")
}
