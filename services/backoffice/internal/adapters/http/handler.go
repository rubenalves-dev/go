package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"raiiaa.dev/services/backoffice/internal/ports"
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
	mux.HandleFunc("POST "+h.adminPrefix+"/admins", h.createAdmin)
	mux.HandleFunc("GET "+h.adminPrefix+"/admins/{id}", h.getAdmin)
	mux.HandleFunc("PUT "+h.adminPrefix+"/admins/{id}", h.updateAdmin)
	mux.HandleFunc("DELETE "+h.adminPrefix+"/admins/{id}", h.deleteAdmin)
	return mux
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) playground(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(fmt.Sprintf(
		backofficePlaygroundHTML,
		h.adminPrefix, h.adminPrefix, h.adminPrefix, h.adminPrefix, h.adminPrefix, h.adminPrefix,
		h.adminPrefix, h.adminPrefix, h.adminPrefix, h.adminPrefix, h.adminPrefix, h.adminPrefix,
	)))
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
		h.handleError(w, err, "failed to list admins")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": admins})
}

type upsertAdminRequest struct {
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

func (h *Handler) createAdmin(w http.ResponseWriter, r *http.Request) {
	var req upsertAdminRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	admin, err := h.adminService.CreateAdmin(r.Context(), req.Email, req.Role, req.IsActive)
	if err != nil {
		h.handleError(w, err, "failed to create admin")
		return
	}

	writeJSON(w, http.StatusCreated, admin)
}

func (h *Handler) getAdmin(w http.ResponseWriter, r *http.Request) {
	admin, err := h.adminService.GetAdminByID(r.Context(), r.PathValue("id"))
	if err != nil {
		h.handleError(w, err, "failed to get admin")
		return
	}
	writeJSON(w, http.StatusOK, admin)
}

func (h *Handler) updateAdmin(w http.ResponseWriter, r *http.Request) {
	var req upsertAdminRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	admin, err := h.adminService.UpdateAdmin(r.Context(), r.PathValue("id"), req.Email, req.Role, req.IsActive)
	if err != nil {
		h.handleError(w, err, "failed to update admin")
		return
	}

	writeJSON(w, http.StatusOK, admin)
}

func (h *Handler) deleteAdmin(w http.ResponseWriter, r *http.Request) {
	if err := h.adminService.DeleteAdmin(r.Context(), r.PathValue("id")); err != nil {
		h.handleError(w, err, "failed to delete admin")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleError(w http.ResponseWriter, err error, logMsg string) {
	switch {
	case errors.Is(err, usecase.ErrInvalidAdminID),
		errors.Is(err, usecase.ErrInvalidAdminEmail),
		errors.Is(err, usecase.ErrInvalidAdminRole):
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
	case errors.Is(err, ports.ErrAdminNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "admin not found"})
	case errors.Is(err, ports.ErrAdminAlreadyExist):
		writeJSON(w, http.StatusConflict, map[string]string{"error": "admin already exists"})
	default:
		h.logger.Error(logMsg, "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func decodeJSON(r *http.Request, out any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
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
    .container { max-width: 980px; }
    h1 { margin-bottom: 0.2rem; }
    .meta { color: #6b7280; margin-bottom: 1.2rem; }
    .card { border: 1px solid #e5e7eb; border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
    .row { display: flex; gap: 0.6rem; flex-wrap: wrap; margin-bottom: 0.6rem; }
    input, select { padding: 0.5rem; border: 1px solid #d1d5db; border-radius: 6px; min-width: 180px; }
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

    <div class="card">
      <h2>Quick checks</h2>
      <div class="row">
        <button id="status-btn">GET %s/status</button>
        <button id="admins-btn">GET %s/admins</button>
      </div>
    </div>

    <div class="card">
      <h2>Create admin (POST %s/admins)</h2>
      <div class="row">
        <input id="create-email" placeholder="admin@example.com">
        <select id="create-role">
          <option value="owner">owner</option>
          <option value="editor">editor</option>
          <option value="viewer">viewer</option>
        </select>
        <select id="create-active">
          <option value="true">active</option>
          <option value="false">inactive</option>
        </select>
        <button id="create-btn">Create</button>
      </div>
    </div>

    <div class="card">
      <h2>Get admin by id (GET %s/admins/{id})</h2>
      <div class="row">
        <input id="get-id" placeholder="admin id">
        <button id="get-btn">Get</button>
      </div>
    </div>

    <div class="card">
      <h2>Update admin (PUT %s/admins/{id})</h2>
      <div class="row">
        <input id="update-id" placeholder="admin id">
        <input id="update-email" placeholder="admin@example.com">
        <select id="update-role">
          <option value="owner">owner</option>
          <option value="editor">editor</option>
          <option value="viewer">viewer</option>
        </select>
        <select id="update-active">
          <option value="true">active</option>
          <option value="false">inactive</option>
        </select>
        <button id="update-btn">Update</button>
      </div>
    </div>

    <div class="card">
      <h2>Delete admin (DELETE %s/admins/{id})</h2>
      <div class="row">
        <input id="delete-id" placeholder="admin id">
        <button id="delete-btn">Delete</button>
      </div>
    </div>

    <div class="card">
      <h2>Output</h2>
      <pre id="out">Ready.</pre>
    </div>
  </div>

  <script>
    const out = document.getElementById("out");

    async function callAPI(path, method = "GET", payload = null) {
      const options = { method };
      if (payload !== null) {
        options.headers = { "Content-Type": "application/json" };
        options.body = JSON.stringify(payload);
      }

      const res = await fetch(path, options);
      const text = await res.text();
      out.textContent = "HTTP " + res.status + "\n" + text;
    }

    function isActive(value) {
      return value === "true";
    }

    document.getElementById("status-btn").addEventListener("click", async () => {
      await callAPI("%s/status");
    });

    document.getElementById("admins-btn").addEventListener("click", async () => {
      await callAPI("%s/admins");
    });

    document.getElementById("create-btn").addEventListener("click", async () => {
      await callAPI("%s/admins", "POST", {
        email: document.getElementById("create-email").value,
        role: document.getElementById("create-role").value,
        is_active: isActive(document.getElementById("create-active").value),
      });
    });

    document.getElementById("get-btn").addEventListener("click", async () => {
      const id = document.getElementById("get-id").value;
      await callAPI("%s/admins/" + encodeURIComponent(id));
    });

    document.getElementById("update-btn").addEventListener("click", async () => {
      const id = document.getElementById("update-id").value;
      await callAPI("%s/admins/" + encodeURIComponent(id), "PUT", {
        email: document.getElementById("update-email").value,
        role: document.getElementById("update-role").value,
        is_active: isActive(document.getElementById("update-active").value),
      });
    });

    document.getElementById("delete-btn").addEventListener("click", async () => {
      const id = document.getElementById("delete-id").value;
      await callAPI("%s/admins/" + encodeURIComponent(id), "DELETE");
    });
  </script>
</body>
</html>`
