package http

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"raiiaa.dev/services/backoffice/internal/usecase"
)

type Handler struct {
	adminService *usecase.AdminService
	adminPrefix  string
	logger       *slog.Logger
}

func NewHandler(adminService *usecase.AdminService, adminPrefix string, logger *slog.Logger) *Handler {
	return &Handler{
		adminService: adminService,
		adminPrefix:  adminPrefix,
		logger:       logger,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.playground)
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET "+h.adminPrefix+"/status", h.adminStatus)
	mux.HandleFunc("GET "+h.adminPrefix+"/admins", h.admins)
	return mux
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) playground(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf(backofficePlaygroundHTML, h.adminPrefix, h.adminPrefix, h.adminPrefix, h.adminPrefix)))
}

func (h *Handler) adminStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.adminService.GetStatus(r.Context())
	if err != nil {
		h.logger.Error("failed to fetch admin status", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (h *Handler) admins(w http.ResponseWriter, r *http.Request) {
	admins, err := h.adminService.ListAdmins(r.Context())
	if err != nil {
		h.logger.Error("failed to list admins", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": admins})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

const backofficePlaygroundHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Backoffice Playground</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 2rem; color: #111827; }
    .container { max-width: 920px; }
    h1 { margin-bottom: 0.2rem; }
    .meta { color: #6b7280; margin-bottom: 1.2rem; }
    .row { display: flex; gap: 0.6rem; flex-wrap: wrap; margin-bottom: 1rem; }
    button { padding: 0.5rem 0.8rem; border: 0; border-radius: 6px; background: #111827; color: #fff; cursor: pointer; }
    pre { background: #0f172a; color: #e2e8f0; padding: 0.8rem; border-radius: 8px; overflow-x: auto; }
    a { color: #1d4ed8; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Backoffice Playground</h1>
    <p class="meta">Service-local UI for testing backoffice endpoints directly.</p>
    <p><a href="http://localhost:8080/">Back to gateway nav</a></p>
    <div class="row">
      <button id="status-btn">GET %s/status</button>
      <button id="admins-btn">GET %s/admins</button>
    </div>
    <pre id="out">Ready.</pre>
  </div>

  <script>
    const out = document.getElementById("out");

    async function callAPI(path) {
      const res = await fetch(path);
      const text = await res.text();
      out.textContent = "HTTP " + res.status + "\n" + text;
    }

    document.getElementById("status-btn").addEventListener("click", async () => {
      await callAPI("%s/status");
    });

    document.getElementById("admins-btn").addEventListener("click", async () => {
      await callAPI("%s/admins");
    });
  </script>
</body>
</html>`
