package httpserver

import (
	"errors"
	"go-platform-template/internal/core/apperror"
	"log/slog"
	"net/http"
)

// Type: machine-readable error kind (snake_case, stable contract);
// HTTP status carries severity, Type carries category
type errorBody struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func newErrorResponse(errorType string, errorMessage string) errorResponse {
	return errorResponse{
		Error: errorBody{
			Type:    errorType,
			Message: errorMessage,
		},
	}
}

func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// package-level slog: SetDefault guarantees our handler; injection overkill for helpers
func writeError(w http.ResponseWriter, err error) {
	var responseStatus int
	var response errorResponse

	switch {
	case errors.Is(err, errBadRequest):
		responseStatus = http.StatusBadRequest
		response = newErrorResponse("bad_request", err.Error())
	case errors.Is(err, apperror.ErrInvalid):
		responseStatus = http.StatusBadRequest
		response = newErrorResponse("invalid_request", err.Error())
	case errors.Is(err, apperror.ErrNotFound):
		responseStatus = http.StatusNotFound
		response = newErrorResponse("not_found", err.Error())
	// Expected internal server error
	case errors.Is(err, apperror.ErrInternal):
		responseStatus = http.StatusInternalServerError
		response = newErrorResponse("internal_error", "internal server error")
	// Unexpected that we need to log
	default:
		responseStatus = http.StatusInternalServerError
		response = newErrorResponse("internal_error", "internal server error")
		slog.Error("returned undefined error to user", "err", err)
	}

	_ = writeJSON(w, responseStatus, response)
}
