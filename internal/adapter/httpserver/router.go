package httpserver

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Router struct {
	reliabilityHandler *ProbeHandler
	widgetHandler      *WidgetHandler
}

func NewRouter(reliabilityHandler *ProbeHandler, widgetHandler *WidgetHandler) *Router {
	return &Router{
		reliabilityHandler: reliabilityHandler,
		widgetHandler:      widgetHandler,
	}
}

func (r *Router) RegisterRoutes() http.Handler {
	mux := http.NewServeMux()

	// Probe & observability
	mux.HandleFunc("GET /healthz", r.reliabilityHandler.Healthz)
	mux.HandleFunc("GET /readyz", r.reliabilityHandler.Readyz)
	// default registry via promauto; switch to an explicit Registry + HandlerFor when isolation matters
	mux.Handle("GET /metrics", promhttp.Handler())

	// Widget
	mux.HandleFunc("POST /v1/widgets", r.widgetHandler.Create)
	mux.HandleFunc("GET /v1/widgets/{id}", r.widgetHandler.Get)
	mux.HandleFunc("PUT /v1/widgets/{id}", r.widgetHandler.Update)
	mux.HandleFunc("DELETE /v1/widgets/{id}", r.widgetHandler.Delete)
	mux.HandleFunc("GET /v1/widgets", r.widgetHandler.List)

	return recoverMiddleware(requestIDMiddleware(observabilityMiddleware(mux)))
}
