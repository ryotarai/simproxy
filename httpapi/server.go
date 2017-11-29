package httpapi

import "net/http"

func Start(addr string, b balancer) {
	s := &http.Server{
		Addr:    addr,
		Handler: NewHandler(b),
	}
	go func() {
		s.ListenAndServe()
	}()
}
