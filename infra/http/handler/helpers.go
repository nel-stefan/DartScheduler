package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"DartScheduler/domain"
)

func writeJSON(w http.ResponseWriter, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("[ERROR] [writeJSON] encode error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

func httpError(w http.ResponseWriter, err error, code int) {
	log.Printf("[ERROR] status=%d err=%v", code, err)
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
	case errors.Is(err, domain.ErrScheduleConstraintViolation):
		code = http.StatusUnprocessableEntity
	default:
		code = http.StatusInternalServerError
	}
	log.Printf("[ERROR] status=%d err=%v", code, err)
	if code >= 500 {
		http.Error(w, "internal server error", code)
	} else {
		http.Error(w, err.Error(), code)
	}
}
