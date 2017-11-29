package httpapi

import (
	"fmt"
	"net/http"

	"github.com/ryotarai/simproxy/types"
)

type balancer interface {
	Metrics() map[*types.Backend]map[string]int64
}

type Handler struct {
	balancer balancer
}

func NewHandler(b balancer) *Handler {
	return &Handler{
		balancer: b,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/metrics" {
		h.handleMetrics(w, r)
	} else {
		w.WriteHeader(404)
		fmt.Fprintln(w, "404 not found")
	}
}

func (h *Handler) handleMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.balancer.Metrics()
	for be, m := range metrics {
		url := be.URL.String()
		fmt.Fprintf(w, "simproxy_backend_weight{backend=%s} %f\n", url, be.Weight)
		for k, v := range m {
			fmt.Fprintf(w, "simproxy_%s{backend=%s} %d\n", k, url, v)
		}
	}
}
