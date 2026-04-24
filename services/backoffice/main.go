package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"raiiaa.dev/common/logger"
	httpadapter "raiiaa.dev/services/backoffice/internal/adapters/http"
	"raiiaa.dev/services/backoffice/internal/adapters/postgres"
	"raiiaa.dev/services/backoffice/internal/config"
	"raiiaa.dev/services/backoffice/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("load config: %w", err))
	}

	log := logger.New(cfg.ServiceName, cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Error("failed creating postgres pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	adminRepo := postgres.NewAdminRepository(pool)
	adminService := usecase.NewAdminService(cfg.ServiceName, usecase.SystemClock{}, adminRepo)
	handler := httpadapter.NewHandler(adminService, cfg.AdminRoutePrefix, log)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("backoffice started", "addr", cfg.HTTPAddr, "admin_prefix", cfg.AdminRoutePrefix)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("http server failed", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}

	log.Info("backoffice stopped")
}
