package simproxy

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ryotarai/simproxy/handler"
)

type Handler struct {
	Balancer           Balancer
	Logger             *log.Logger
	AccessLogger       handler.AccessLogger
	BackendURLHeader   string
	Transport          *http.Transport
	EnableBackendTrace bool

	handler http.Handler
}

func (h *Handler) Setup() {
	transport := http.DefaultTransport
	if h.Transport != nil {
		transport = h.Transport
	}

	h.handler = &handler.ReverseProxy{
		AccessLogger:      h.AccessLogger,
		ErrorLog:          h.Logger,
		BackendURLHeader:  h.BackendURLHeader,
		Transport:         transport,
		EnableClientTrace: h.EnableBackendTrace,

		PickBackend:    h.pickBackend,
		AfterRoundTrip: h.afterRoundTrip,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}

func (h *Handler) pickBackend() (handler.Backend, error) {
	backend, err := h.Balancer.PickBackend()
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
	h.Balancer.ReturnBackend(b2)
}
