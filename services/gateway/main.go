package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"raiiaa.dev/common/logger"
	"raiiaa.dev/services/gateway/internal/config"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.ServiceName, cfg.LogLevel)

	authTarget, err := url.Parse(cfg.AuthServiceURL)
	if err != nil {
		log.Error("invalid auth service URL", "error", err, "value", cfg.AuthServiceURL)
		os.Exit(1)
	}

	backofficeTarget, err := url.Parse(cfg.BackofficeServiceURL)
	if err != nil {
		log.Error("invalid backoffice service URL", "error", err, "value", cfg.BackofficeServiceURL)
		os.Exit(1)
	}

	authProxy := newReverseProxy(authTarget, log, "auth-service")
	backofficeProxy := newReverseProxy(backofficeTarget, log, "backoffice")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /api/v1/auth/signup", authProxy.ServeHTTP)
	mux.HandleFunc("POST /api/v1/auth/login", authProxy.ServeHTTP)
	mux.Handle("/api/v1/admin/", backofficeProxy)

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info(
			"gateway started",
			"addr", cfg.HTTPAddr,
			"auth_service_url", cfg.AuthServiceURL,
			"backoffice_service_url", cfg.BackofficeServiceURL,
		)
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

	log.Info("gateway stopped")
}

func newReverseProxy(target *url.URL, log *slog.Logger, upstreamName string) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	proxy.ErrorHandler = func(rw http.ResponseWriter, _ *http.Request, err error) {
		log.Error("proxy error", "upstream", upstreamName, "error", err)
		http.Error(rw, "upstream unavailable", http.StatusBadGateway)
	}
	return proxy
}
