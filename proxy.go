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

	serverstarter "github.com/lestrrat/go-server-starter/listener"
	"github.com/ryotarai/simproxy/handler"
)

type Proxy struct {
	balancer     Balancer
	accessLogger handler.AccessLogger
}

func NewProxy(balancer Balancer, l handler.AccessLogger) *Proxy {
	return &Proxy{
		balancer:     balancer,
		accessLogger: l,
	}
}

func (p *Proxy) ListenAndServe(listen string) error {
	var l net.Listener
	if listen == "SERVER_STARTER" {
		ls, err := serverstarter.ListenAll()
		if err != nil {
			return err
		}
		if len(ls) > 1 {
			return fmt.Errorf("%d sockets (more than 1) are passed by server-starter", len(ls))
		}
		l = ls[0]
	} else {
		var err error
		l, err = net.Listen("tcp", listen)
		if err != nil {
			return err
		}
	}

	defer l.Close()

	return p.Serve(l)
}

func (p *Proxy) Serve(listener net.Listener) error {
	handler := &handler.ReverseProxy{
		AccessLogger:   p.accessLogger,
		Director:       p.director,
		PickBackend:    p.pickBackend,
		AfterRoundTrip: p.afterRoundTrip,
	}
	server := http.Server{
		Handler: handler,
	}

	go func() {
		server.Serve(listener)
	}()

	p.waitSignal(server)

	return nil
}

func (p *Proxy) waitSignal(server http.Server) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
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
