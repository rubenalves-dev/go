package http

import (
	"encoding/json"
	"net/http"

	"raiiaa.dev/services/backoffice/internal/usecase"
)

type Handler struct {
	adminService *usecase.AdminService
	adminPrefix  string
}

func NewHandler(adminService *usecase.AdminService, adminPrefix string) *Handler {
	return &Handler{
		adminService: adminService,
		adminPrefix:  adminPrefix,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("GET "+h.adminPrefix+"/status", h.adminStatus)
	return mux
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) adminStatus(w http.ResponseWriter, r *http.Request) {
	status := h.adminService.GetStatus(r.Context())
	writeJSON(w, http.StatusOK, status)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "failed encoding response", http.StatusInternalServerError)
	}
}
