package main

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
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
	navPageTemplate := template.Must(template.New("service-nav").Parse(serviceNavPageHTML))
	statusPageTemplate := template.Must(template.New("service-status").Parse(serviceStatusPageHTML))
	statusClient := &http.Client{
		Timeout: 3 * time.Second,
	}
	authBrowserURL := "http://localhost:8081/"
	if port := authTarget.Port(); port != "" {
		authBrowserURL = "http://localhost:" + port + "/"
	}
	backofficeBrowserURL := "http://localhost:8082/"
	if port := backofficeTarget.Port(); port != "" {
		backofficeBrowserURL = "http://localhost:" + port + "/"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := navPageTemplate.Execute(w, serviceNavPageData{
			AuthServiceURL:       authBrowserURL,
			BackofficeServiceURL: backofficeBrowserURL,
		}); err != nil {
			log.Error("failed to render navigation page", "error", err)
		}
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /status/services", func(w http.ResponseWriter, r *http.Request) {
		statuses := collectServiceStatuses(r.Context(), statusClient, []serviceTarget{
			{Name: "auth-service", BaseURL: authTarget},
			{Name: "backoffice", BaseURL: backofficeTarget},
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(statuses); err != nil {
			log.Error("failed to encode service statuses", "error", err)
		}
	})
	mux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		statuses := collectServiceStatuses(r.Context(), statusClient, []serviceTarget{
			{Name: "auth-service", BaseURL: authTarget},
			{Name: "backoffice", BaseURL: backofficeTarget},
		})

		healthyCount := 0
		for _, status := range statuses {
			if status.Healthy {
				healthyCount++
			}
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := statusPageTemplate.Execute(w, serviceStatusPageData{
			Statuses:       statuses,
			LastUpdatedUTC: time.Now().UTC().Format(time.RFC3339),
			HealthyCount:   healthyCount,
			TotalCount:     len(statuses),
		}); err != nil {
			log.Error("failed to render service status page", "error", err)
		}
	})
	mux.HandleFunc("POST /api/v1/auth/signup", authProxy.ServeHTTP)
	mux.HandleFunc("POST /api/v1/auth/login", authProxy.ServeHTTP)
	mux.Handle("GET /api/v1/admin/", backofficeProxy)

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

type serviceTarget struct {
	Name    string
	BaseURL *url.URL
}

type serviceStatus struct {
	Name       string `json:"name"`
	HealthURL  string `json:"health_url"`
	Healthy    bool   `json:"healthy"`
	StatusCode int    `json:"status_code"`
	LatencyMS  int64  `json:"latency_ms"`
	Error      string `json:"error,omitempty"`
}

type serviceStatusPageData struct {
	Statuses       []serviceStatus
	LastUpdatedUTC string
	HealthyCount   int
	TotalCount     int
}

type serviceNavPageData struct {
	AuthServiceURL       string
	BackofficeServiceURL string
}

func collectServiceStatuses(ctx context.Context, client *http.Client, services []serviceTarget) []serviceStatus {
	statuses := make([]serviceStatus, 0, len(services))
	for _, service := range services {
		statuses = append(statuses, checkServiceHealth(ctx, client, service))
	}
	return statuses
}

func checkServiceHealth(ctx context.Context, client *http.Client, service serviceTarget) serviceStatus {
	healthURL := service.BaseURL.ResolveReference(&url.URL{Path: "/healthz"}).String()
	status := serviceStatus{
		Name:      service.Name,
		HealthURL: healthURL,
	}

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, healthURL, nil)
	if err != nil {
		status.Error = err.Error()
		return status
	}

	start := time.Now()
	resp, err := client.Do(req)
	status.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		status.Error = err.Error()
		return status
	}
	defer resp.Body.Close()

	status.StatusCode = resp.StatusCode
	status.Healthy = resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
	if !status.Healthy {
		status.Error = resp.Status
	}
	return status
}

const serviceStatusPageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="refresh" content="5">
  <title>Gateway Service Status</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 2rem; color: #111827; }
    table { border-collapse: collapse; width: 100%; max-width: 900px; margin-top: 1rem; }
    th, td { border: 1px solid #e5e7eb; padding: 0.6rem 0.8rem; text-align: left; }
    th { background: #f9fafb; }
    .ok { color: #047857; font-weight: 600; }
    .down { color: #b91c1c; font-weight: 600; }
    .meta { color: #6b7280; margin-top: 0.4rem; }
    code { background: #f3f4f6; padding: 0.1rem 0.3rem; border-radius: 4px; }
  </style>
</head>
<body>
  <h1>Connected Services</h1>
  <p>{{.HealthyCount}} / {{.TotalCount}} services healthy</p>
  <p class="meta">Last updated (UTC): {{.LastUpdatedUTC}} · Auto-refresh: every 5s · JSON: <code>/status/services</code></p>
  <table>
    <thead>
      <tr>
        <th>Service</th>
        <th>Health endpoint</th>
        <th>Status</th>
        <th>HTTP code</th>
        <th>Latency</th>
        <th>Error</th>
      </tr>
    </thead>
    <tbody>
      {{range .Statuses}}
      <tr>
        <td>{{.Name}}</td>
        <td><code>{{.HealthURL}}</code></td>
        <td>{{if .Healthy}}<span class="ok">UP</span>{{else}}<span class="down">DOWN</span>{{end}}</td>
        <td>{{if .StatusCode}}{{.StatusCode}}{{else}}-{{end}}</td>
        <td>{{.LatencyMS}}ms</td>
        <td>{{if .Error}}{{.Error}}{{else}}-{{end}}</td>
      </tr>
      {{end}}
    </tbody>
  </table>
</body>
</html>`

const serviceNavPageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Gateway Navigation</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 2rem; color: #111827; }
    h1 { margin-bottom: 0.4rem; }
    .meta { color: #6b7280; margin-bottom: 1.2rem; }
    ul { padding-left: 1.2rem; }
    li { margin: 0.5rem 0; }
    a { color: #1d4ed8; text-decoration: none; }
    a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <h1>Gateway Navigation</h1>
  <p class="meta">Use this page to jump to service-local playgrounds.</p>
  <ul>
    <li><a href="/status">Gateway service status page</a></li>
    <li><a href="{{.AuthServiceURL}}">Auth service playground</a></li>
    <li><a href="{{.BackofficeServiceURL}}">Backoffice service playground</a></li>
  </ul>
</body>
</html>`
