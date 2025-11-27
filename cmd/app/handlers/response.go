package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	appErr "educabot.com/bookshop/internal/platform/errors"
)

type errorResponse struct {
	Message string `json:"message"`
}

func writeError(w http.ResponseWriter, err error) {

	switch {
	case errors.Is(err, appErr.ErrTimeout):
		writeJSON(w, http.StatusGatewayTimeout, "external service timeout")
		return

	case errors.Is(err, appErr.ErrExternalService):
		writeJSON(w, http.StatusBadGateway, "external service error")
		return

	case errors.Is(err, appErr.ErrInvalidResponse):
		writeJSON(w, http.StatusUnprocessableEntity, "invalid response from external API")
		return

	default:
		writeJSON(w, http.StatusInternalServerError, "internal server error")
		return
	}
}

func writeJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(errorResponse{Message: message})
}
