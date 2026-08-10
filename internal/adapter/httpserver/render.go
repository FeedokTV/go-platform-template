package httpserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

var errBadRequest = errors.New("bad request")

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(dst)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		var syntaxErr *json.SyntaxError

		switch {
		case errors.Is(err, io.EOF):
			return fmt.Errorf("%w: empty body", errBadRequest)
		case errors.As(err, &maxBytesErr):
			return fmt.Errorf("%w: body too large", errBadRequest)
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("%w: malformed JSON", errBadRequest)
		default:
			return fmt.Errorf("%w: invalid request", errBadRequest)
		}
	}

	err = decoder.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: body must only contain a single JSON value", errBadRequest)
	}

	return nil
}

// package-level slog: SetDefault guarantees our handler; injection overkill for helpers
func writeJSON(w http.ResponseWriter, status int, input any) error {
	// First: marshal. So if marshal fails, we can break connection. If we do that later
	// we cannot unsend http.StatusOK and send http.StatusInternalServerError

	data, err := json.Marshal(input)
	if err != nil {
		slog.Error("error while marshalling data", "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("error marshal data: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_, err = w.Write(data)
	if err != nil {
		slog.Debug("error while writing data to response", "err", err)
		return fmt.Errorf("error writing data: %w", err)
	}

	return nil
}
