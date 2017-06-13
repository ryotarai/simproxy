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

	"fmt"

	"github.com/ryotarai/simproxy/handler"
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
	handler := &handler.ReverseProxy{
		AccessLogger:   p,
		Director:       p.director,
		PickBackend:    p.pickBackend,
		AfterRoundTrip: p.afterRoundTrip,
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

func (p *Proxy) director(req *http.Request, backend handler.Backend) {
	handler.StandardDirector(req, backend.GetURL())
}

func (p *Proxy) pickBackend() (handler.Backend, error) {
	backend, err := p.balancer.PickBackend()
	if err != nil {
		return nil, err
	}
	return backend, nil
}

func (p *Proxy) afterRoundTrip(b handler.Backend) {
	b2, ok := b.(*Backend)
	if !ok {
		panic(fmt.Sprintf("%#v is not Backend", b2))
	}
	p.balancer.ReturnBackend(b2)
}

func (p *Proxy) Log(r handler.LogRecord) error {
	log.Println(r)
	return nil
}
