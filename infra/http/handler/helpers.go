package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"DartScheduler/domain"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func httpError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}

func httpErrorDomain(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrInvalidInput):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, domain.ErrAlreadyExists):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, domain.ErrMatchAlreadyPlayed):
		http.Error(w, err.Error(), http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
