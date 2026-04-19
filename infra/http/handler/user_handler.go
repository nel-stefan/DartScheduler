package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"DartScheduler/domain"
	mw "DartScheduler/infra/http/middleware"
	"DartScheduler/usecase"

	"github.com/go-chi/chi/v5"
)

// UserHandler handles user management endpoints (admin only).
type UserHandler struct {
	authUC *usecase.AuthUseCase
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(authUC *usecase.AuthUseCase) *UserHandler {
	return &UserHandler{authUC: authUC}
}

// List handles GET /api/users.
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.authUC.ListUsers(r.Context())
	if err != nil {
		httpError(w, err, http.StatusInternalServerError)
		return
	}
	writeJSON(w, users)
}

// Create handles POST /api/users.
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	dto, err := h.authUC.CreateUser(r.Context(), usecase.CreateUserInput{
		Username: body.Username,
		Password: body.Password,
		Role:     body.Role,
	})
	if err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, dto)
}

// Update handles PUT /api/users/:id — updates role and/or password.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Role     string `json:"role"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	if err := h.authUC.UpdateUser(r.Context(), id, usecase.UpdateUserInput{
		Role:     body.Role,
		Password: body.Password,
	}); err != nil {
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/users/:id.
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	identity, _ := mw.IdentityFromContext(r.Context())
	if err := h.authUC.DeleteUser(r.Context(), id, identity.UserID); err != nil {
		if errors.Is(err, domain.ErrInvalidInput) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		httpErrorDomain(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
