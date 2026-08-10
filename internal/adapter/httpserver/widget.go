package httpserver

import (
	"fmt"
	"go-platform-template/internal/core/widget"
	"net/http"
	"strconv"
	"time"
)

type WidgetHandler struct {
	svc *widget.Service
}

func NewWidgetHandler(svc *widget.Service) *WidgetHandler {
	return &WidgetHandler{svc: svc}
}

type widgetResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Weight    int       `json:"weight"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toWidgetResponse(w widget.Widget) widgetResponse {
	return widgetResponse{
		ID:        w.ID,
		Name:      w.Name,
		Weight:    w.Weight,
		CreatedAt: w.CreatedAt,
		UpdatedAt: w.UpdatedAt,
	}
}

type createWidgetRequest struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

type updateWidgetRequest struct {
	Name   string `json:"name"`
	Weight int    `json:"weight"`
}

func (h *WidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var userInput createWidgetRequest

	err := decodeJSON(w, r, &userInput)
	if err != nil {
		writeError(w, err)
		return
	}

	createdWidget, err := h.svc.Create(r.Context(), widget.Widget{Name: userInput.Name, Weight: userInput.Weight})
	if err != nil {
		writeError(w, err)
		return
	}

	// Success
	_ = writeJSON(w, http.StatusCreated, toWidgetResponse(createdWidget))

}

func (h *WidgetHandler) Get(w http.ResponseWriter, r *http.Request) {
	widgetIDStr := r.PathValue("id")
	widgetID, err := strconv.ParseInt(widgetIDStr, 10, 64)
	if err != nil {
		writeError(w, fmt.Errorf("%w: invalid id", errBadRequest))
		return
	}

	fetchedWidget, err := h.svc.Get(r.Context(), widgetID)
	if err != nil {
		writeError(w, err)
		return
	}

	// Success
	_ = writeJSON(w, http.StatusOK, toWidgetResponse(fetchedWidget))
}

func (h *WidgetHandler) Update(w http.ResponseWriter, r *http.Request) {
	widgetIDStr := r.PathValue("id")
	widgetID, err := strconv.ParseInt(widgetIDStr, 10, 64)
	if err != nil {
		writeError(w, fmt.Errorf("%w: invalid id", errBadRequest))
		return
	}

	var userInput updateWidgetRequest
	err = decodeJSON(w, r, &userInput)
	if err != nil {
		writeError(w, err)
		return
	}

	fetchedWidget, err := h.svc.Update(r.Context(), widget.Widget{
		ID:     widgetID,
		Name:   userInput.Name,
		Weight: userInput.Weight,
	})

	if err != nil {
		writeError(w, err)
		return
	}

	// Success
	_ = writeJSON(w, http.StatusOK, toWidgetResponse(fetchedWidget))
}

func (h *WidgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	widgetIDStr := r.PathValue("id")
	widgetID, err := strconv.ParseInt(widgetIDStr, 10, 64)
	if err != nil {
		writeError(w, fmt.Errorf("%w: invalid id", errBadRequest))
		return
	}

	err = h.svc.Delete(r.Context(), widgetID)
	if err != nil {
		writeError(w, err)
		return
	}

	// Success
	writeNoContent(w)
}

func (h *WidgetHandler) List(w http.ResponseWriter, r *http.Request) {

	widgets, err := h.svc.List(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	widgetResponses := []widgetResponse{}

	for _, widget := range widgets {
		widgetResponses = append(widgetResponses, toWidgetResponse(widget))
	}

	// Success
	_ = writeJSON(w, http.StatusOK, widgetResponses)
}
