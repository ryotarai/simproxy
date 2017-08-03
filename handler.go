package simproxy

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ryotarai/simproxy/balancer"
	"github.com/ryotarai/simproxy/handler"
	"github.com/ryotarai/simproxy/types"
)

type Handler struct {
	Balancer            balancer.Balancer
	Logger              *log.Logger
	AccessLogger        handler.AccessLogger
	BackendURLHeader    string
	Transport           *http.Transport
	EnableBackendTrace  bool
	AppendXForwardedFor bool
	EnableBufferPool    bool

	handler http.Handler
}

func (h *Handler) Setup() {
	transport := http.DefaultTransport
	if h.Transport != nil {
		transport = h.Transport
	}

	var bufferPool handler.BufferPool
	if h.EnableBufferPool {
		h.Logger.Println("INFO: using buffer pool")
		bufferPool = NewBufferPool(32 * 1024)
	}

	h.handler = &handler.ReverseProxy{
		AccessLogger:        h.AccessLogger,
		ErrorLog:            h.Logger,
		BackendURLHeader:    h.BackendURLHeader,
		Transport:           transport,
		EnableClientTrace:   h.EnableBackendTrace,
		AppendXForwardedFor: h.AppendXForwardedFor,
		BufferPool:          bufferPool,

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
	b2, ok := b.(*types.Backend)
	if !ok {
		panic(fmt.Sprintf("%#v is not Backend", b2))
	}
	h.Balancer.ReturnBackend(b2)
}
