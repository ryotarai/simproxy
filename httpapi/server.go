package httpapi

import (
	"net"
	"net/http"
)

func Start(l net.Listener, b balancer) {
	s := &http.Server{
		Handler: NewHandler(b),
	}
	go func() {
		s.Serve(l)
	}()
}
