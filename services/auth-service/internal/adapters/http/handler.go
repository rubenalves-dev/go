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
	mux.HandleFunc("GET /", h.playground)
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

func (h *Handler) playground(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(authPlaygroundHTML))
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

const authPlaygroundHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Auth Service Playground</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 2rem; color: #111827; }
    .container { max-width: 920px; }
    h1 { margin-bottom: 0.2rem; }
    .meta { color: #6b7280; margin-bottom: 1.2rem; }
    .card { border: 1px solid #e5e7eb; border-radius: 8px; padding: 1rem; margin-bottom: 1rem; }
    .row { display: flex; gap: 0.6rem; flex-wrap: wrap; }
    input { padding: 0.5rem; border: 1px solid #d1d5db; border-radius: 6px; min-width: 220px; }
    button { padding: 0.5rem 0.8rem; border: 0; border-radius: 6px; background: #111827; color: #fff; cursor: pointer; }
    pre { background: #0f172a; color: #e2e8f0; padding: 0.8rem; border-radius: 8px; overflow-x: auto; }
    a { color: #1d4ed8; }
  </style>
</head>
<body>
  <div class="container">
    <h1>Auth Service Playground</h1>
    <p class="meta">Service-local UI for testing signup/login directly on auth-service.</p>
    <p><a href="http://localhost:8080/">Back to gateway nav</a></p>

    <div class="card">
      <h2>Signup</h2>
      <div class="row">
        <input id="signup-email" placeholder="email@example.com">
        <input id="signup-password" type="password" placeholder="password">
        <button id="signup-btn">POST /api/v1/auth/signup</button>
      </div>
    </div>

    <div class="card">
      <h2>Login</h2>
      <div class="row">
        <input id="login-email" placeholder="email@example.com">
        <input id="login-password" type="password" placeholder="password">
        <button id="login-btn">POST /api/v1/auth/login</button>
      </div>
    </div>

    <div class="card">
      <h2>Output</h2>
      <pre id="out">Ready.</pre>
    </div>
  </div>

  <script>
    const out = document.getElementById("out");

    async function callAuth(path, payload) {
      const res = await fetch(path, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      const text = await res.text();
      out.textContent = "HTTP " + res.status + "\n" + text;
    }

    document.getElementById("signup-btn").addEventListener("click", async () => {
      await callAuth("/api/v1/auth/signup", {
        email: document.getElementById("signup-email").value,
        password: document.getElementById("signup-password").value,
      });
    });

    document.getElementById("login-btn").addEventListener("click", async () => {
      await callAuth("/api/v1/auth/login", {
        email: document.getElementById("login-email").value,
        password: document.getElementById("login-password").value,
      });
    });
  </script>
</body>
</html>`
