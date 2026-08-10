package httpserver

import (
	"context"
	"net/http"
	"time"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type ProbeHandler struct {
	pinger Pinger
}

func NewProbeHandler(pinger Pinger) *ProbeHandler {
	return &ProbeHandler{pinger: pinger}
}

func (rh *ProbeHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	err := rh.pinger.Ping(pingCtx)
	if err != nil {
		_ = writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "unavailable"})
		return
	}

	_ = writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}

func (rh *ProbeHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	_ = writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
}
