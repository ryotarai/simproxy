package simproxy

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryotarai/simproxy/httputil"
)

type Proxy struct {
	balancer Balancer
}

func NewProxy(balancer Balancer) *Proxy {
	return &Proxy{
		balancer: balancer,
	}
}

func (p *Proxy) Serve(listen string) error {
	handler := &httputil.ReverseProxy{
		AccessLogger: p,
		Director:     p.director,
	}
	server := http.Server{
		Handler: handler,
	}

	l, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	go func() {
		defer l.Close()
		server.Serve(l)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	return nil
}

func (p *Proxy) director(req *http.Request) (func(), string, error) {
	backend, err := p.balancer.PickBackend()
	if err != nil {
		return nil, "", err
	}
	httputil.StandardDirector(req, backend.URL)

	return func() {
		p.balancer.ReturnBackend(backend)
	}, backend.URL.String(), nil
}

func (p *Proxy) Log(r httputil.LogRecord) error {
	log.Println(r)
	return nil
}
