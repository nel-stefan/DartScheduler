package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"DartScheduler/domain"
	mw "DartScheduler/infra/http/middleware"
	"DartScheduler/usecase"
)

// AuthHandler handles login and identity probe requests.
type AuthHandler struct {
	authUC    *usecase.AuthUseCase
	jwtSecret string
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authUC *usecase.AuthUseCase, jwtSecret string) *AuthHandler {
	return &AuthHandler{authUC: authUC, jwtSecret: jwtSecret}
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	out, err := h.authUC.Login(r.Context(), body.Username, body.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, map[string]string{
		"token":    out.Token,
		"username": out.Username,
		"role":     out.Role,
	})
}

// Me handles GET /api/auth/me — resolves identity from network trust or JWT without middleware.
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	identity, ok := mw.ResolveIdentity(r, h.jwtSecret)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, map[string]string{
		"username": identity.Username,
		"role":     identity.Role,
	})
}
