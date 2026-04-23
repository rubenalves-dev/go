package http

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"raiiaa.dev/services/auth-service/internal/ports"
	"raiiaa.dev/services/auth-service/internal/usecase"
)

type Handler struct {
	authService *usecase.AuthService
	logger      *slog.Logger
}

func NewHandler(authService *usecase.AuthService, logger *slog.Logger) *Handler {
	return &Handler{
		authService: authService,
		logger:      logger,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", h.healthz)
	mux.HandleFunc("POST /api/v1/auth/signup", h.signup)
	mux.HandleFunc("POST /api/v1/auth/login", h.login)
	return mux
}

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authResponse struct {
	Token string `json:"token"`
	User  user   `json:"user"`
}

type user struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
}

func (h *Handler) signup(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	token, newUser, err := h.authService.Signup(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ports.ErrUserAlreadyExist):
			writeJSON(w, http.StatusConflict, map[string]string{"error": "user already exists"})
		default:
			h.logger.Error("signup failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{
		Token: token,
		User: user{
			ID:    newUser.ID,
			Email: newUser.Email,
		},
	})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := decodeJSON(r, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	token, existingUser, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidCredentials):
			writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
		default:
			h.logger.Error("login failed", "error", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		}
		return
	}

	writeJSON(w, http.StatusOK, authResponse{
		Token: token,
		User: user{
			ID:    existingUser.ID,
			Email: existingUser.Email,
		},
	})
}

func (h *Handler) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
