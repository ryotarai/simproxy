package simproxy

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/ryotarai/simproxy/handler"
)

type Handler struct {
	handler  http.Handler
	balancer Balancer
}

func NewHandler(balancer Balancer, logger *log.Logger, accessLogger handler.AccessLogger, backendURLHeader string, maxIdleConns int, maxIdleConnsPerHost int) *Handler {
	h := &Handler{
		balancer: balancer,
	}

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          maxIdleConns,
		MaxIdleConnsPerHost:   maxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	h.handler = &handler.ReverseProxy{
		AccessLogger:     accessLogger,
		PickBackend:      h.pickBackend,
		AfterRoundTrip:   h.afterRoundTrip,
		ErrorLog:         logger,
		BackendURLHeader: backendURLHeader,
		Transport:        transport,
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
