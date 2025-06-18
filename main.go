package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"nickbar_v2/internal/config"
	"nickbar_v2/internal/storage/pgsql"
	"os"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error in config: %s", err)
	}

	// TODO: Delete later
	fmt.Println(cfg)

	// APP_ENV as in config.go - local/prod
	log := setupLogger(os.Getenv("APP_ENV"))

	log.Info("starting nickbar")
	log.Debug("debug messages are enabled")

	// TODO: init storage - PostgreSQL
	_, err = pgsql.NewPgxPool(ctx)

	// TODO: init router - gorilla/mux

	// TODO: run server

}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
