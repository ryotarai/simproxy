package simproxy

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ryotarai/simproxy/handler"
)

type Handler struct {
	handler  http.Handler
	balancer Balancer
}

func NewHandler(balancer Balancer, logger *log.Logger, accessLogger handler.AccessLogger) *Handler {
	h := &Handler{
		balancer: balancer,
	}

	h.handler = &handler.ReverseProxy{
		AccessLogger:   accessLogger,
		PickBackend:    h.pickBackend,
		AfterRoundTrip: h.afterRoundTrip,
		ErrorLog:       logger,
	}

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *Handler) pickBackend() (handler.Backend, error) {
	backend, err := h.balancer.PickBackend()
	if err != nil {
		return nil, err
	}
	return backend, nil
}

func (h *Handler) afterRoundTrip(b handler.Backend) {
	b2, ok := b.(*Backend)
	if !ok {
		panic(fmt.Sprintf("%#v is not Backend", b2))
	}
	h.balancer.ReturnBackend(b2)
}
