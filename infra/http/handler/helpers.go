package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"DartScheduler/domain"
)

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[writeJSON] encode error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func httpError(w http.ResponseWriter, err error, code int) {
	log.Printf("[httpError] status=%d err=%v", code, err)
	http.Error(w, err.Error(), code)
}

func httpErrorDomain(w http.ResponseWriter, err error) {
	var code int
	switch {
	case errors.Is(err, domain.ErrNotFound):
		code = http.StatusNotFound
	case errors.Is(err, domain.ErrInvalidInput):
		code = http.StatusBadRequest
	case errors.Is(err, domain.ErrAlreadyExists):
		code = http.StatusConflict
	case errors.Is(err, domain.ErrMatchAlreadyPlayed):
		code = http.StatusConflict
	default:
		code = http.StatusInternalServerError
	}
	log.Printf("[httpErrorDomain] status=%d err=%v", code, err)
	http.Error(w, err.Error(), code)
}
