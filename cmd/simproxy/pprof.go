package main

import (
	"net/http"
	"net/http/pprof"
)

func startPprofServer(addr string) {
	m := http.NewServeMux()
	m.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	m.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	m.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	m.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	m.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	s := &http.Server{
		Addr:    addr,
		Handler: m,
	}
	go func() {
		s.ListenAndServe()
	}()
}
