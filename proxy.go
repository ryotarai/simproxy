package simproxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strings"
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
		Director:       p.director,
		WriteAccessLog: p.writeAccessLog,
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

func (p *Proxy) director(req *http.Request, lr httputil.LogRecord) func() {
	backend := p.balancer.PickBackend()
	lr["backend"] = backend.URL.String()
	httputil.StandardDirector(req, backend.URL)

	return func() {
		p.balancer.ReturnBackend(backend)
	}
}

func (p *Proxy) writeAccessLog(lr httputil.LogRecord) {
	keys := []string{}
	for k := range lr {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fields := []string{}
	for _, k := range keys {
		fields = append(fields, fmt.Sprintf("%s:%s", k, lr[k]))
	}
	line := strings.Join(fields, "\t")

	log.Println(line)
}
