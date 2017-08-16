package simproxy

import (
	"log"
	"net/http"

	"github.com/ryotarai/simproxy/balancer"
	"github.com/ryotarai/simproxy/bufferpool"
	"github.com/ryotarai/simproxy/handler"
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
		bufferPool = bufferpool.New(32 * 1024)
	}

	h.handler = &handler.ReverseProxy{
		AccessLogger:        h.AccessLogger,
		ErrorLog:            h.Logger,
		BackendURLHeader:    h.BackendURLHeader,
		Transport:           transport,
		EnableClientTrace:   h.EnableBackendTrace,
		AppendXForwardedFor: h.AppendXForwardedFor,
		BufferPool:          bufferPool,
		Balancer:            h.Balancer,
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.handler.ServeHTTP(w, r)
}
