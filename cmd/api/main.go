package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/pirellik/sequence-api/internal/config"
	"github.com/pirellik/sequence-api/internal/db"
	"github.com/pirellik/sequence-api/internal/sequence"
	"github.com/pirellik/sequence-api/internal/server"
	"github.com/pirellik/sequence-api/pkg/logger"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load(nil)
	if err != nil {
		log.Fatal(ctx, "loading config", "err", err)
	}

	logger := logger.New(
		cfg.Logger.SlogLevel(),
		cfg.Logger.HumanReadable,
	)
	slog.SetDefault(logger)

	dbPool, err := db.New(ctx, cfg.DB.URL())
	if err != nil {
		slog.ErrorContext(ctx, "initializing db", "err", err)
		os.Exit(1)
	}
	defer dbPool.Close()

	slog.InfoContext(ctx, "migrating up")
	if err := db.MigrateUp(cfg.DB.URL()); err != nil {
		slog.ErrorContext(ctx, "migrating db", "err", err)
		os.Exit(1)
	}

	seqService := sequence.NewService(dbPool)
	handler := server.NewHandler(seqService)
	srv := server.New(handler, cfg.API.Port)

	go func() {
		slog.InfoContext(ctx, "starting api server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.ErrorContext(ctx, "starting api server", "err", err)
			os.Exit(1)
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan
	slog.InfoContext(ctx, "shutting down the API")

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.ErrorContext(ctx, "shutting down the API", "err", err)
	}
}
