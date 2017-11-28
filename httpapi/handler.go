package httpapi

import (
	"net/http"
)

type Handler struct {
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/metrics" {
		h.handleMetrics(w, r)
	}
	w.WriteHeader(404)
}

func (h *Handler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/metrics\n"))
}
