package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"go-platform-template/internal/core/apperror"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type ctxKey int

const requestIDKey ctxKey = 0
const RequestIDHeader = "X-Request-ID"

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed.",
		},
		[]string{"pattern", "method", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of response latencies for HTTP requests.",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"pattern", "method"},
	)
)

func getRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func generateRequestID() string {
	randomBytes := make([]byte, 16)
	// crypto/rand.Read never returns an error as of Go 1.24 (panics on catastrophic entropy failure)
	_, _ = rand.Read(randomBytes)
	return hex.EncodeToString(randomBytes)
}

func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)

		w.Header().Set(RequestIDHeader, requestID)
		next.ServeHTTP(w, r)
	})
}

func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// Re-panic the ErrAbortHandler
			if p := recover(); p != nil {
				if p == http.ErrAbortHandler { //nolint:errorlint
					panic(p)
				}

				// Connection state is unknown after panic; force close
				w.Header().Set("Connection", "close")

				// upstream of the ID middleware, the request context predates the ID
				requestID := w.Header().Get(RequestIDHeader)

				slog.Error("panic recovered", "requestID", requestID, "panic", p, "stack", string(debug.Stack()))
				writeError(w, apperror.ErrInternal)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// Logging middleware
// wrapper hides optional interfaces (Flusher/Hijacker);
// acceptable for JSON APIs; revisit for streaming/websockets, or use http.ResponseController
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.status = http.StatusOK
	}
	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n
	return n, err
}

func observabilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestStartTime := time.Now()

		recorder := &statusRecorder{ResponseWriter: w}
		defer func() {
			slog.Info("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"pattern", r.Pattern,
				"status", recorder.status,
				"bytes", recorder.bytes,
				"duration", time.Since(requestStartTime),
				"requestID", getRequestID(r.Context()))

			duration := time.Since(requestStartTime).Seconds()
			statusCodeStr := strconv.Itoa(recorder.status)
			pattern := r.Pattern
			if pattern == "" {
				pattern = "unmatched"
			}

			httpRequestsTotal.WithLabelValues(pattern, r.Method, statusCodeStr).Inc()
			httpRequestDuration.WithLabelValues(pattern, r.Method).Observe(duration)

		}()
		next.ServeHTTP(recorder, r)
	})
}
