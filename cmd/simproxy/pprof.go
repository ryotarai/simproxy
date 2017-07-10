package main

import (
	"net/http"
	"net/http/pprof"
)

func startPprofServer(addr string) {
	s := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(pprof.Index),
	}
	go func() {
		s.ListenAndServe()
	}()
}
