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
)

type Proxy struct {
	Logger            *log.Logger
	Handler           http.Handler
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	ShutdownTimeout   time.Duration
}

func NewProxy(handler http.Handler, logger *log.Logger) *Proxy {
	return &Proxy{
		Logger:  logger,
		Handler: handler,
	}
}

func (p *Proxy) ListenAndServe(listen string) error {
	l, err := listenFromDescription(listen)
	if err != nil {
		return err
	}
	p.Logger.Printf("listening on %s", l.Addr())
	defer l.Close()

	return p.Serve(l) // block
}

func (p *Proxy) Serve(listener net.Listener) error {
	server := http.Server{
		ErrorLog:          p.Logger,
		Handler:           p.Handler,
		ReadTimeout:       p.ReadTimeout,
		ReadHeaderTimeout: p.ReadHeaderTimeout,
		WriteTimeout:      p.WriteTimeout,
	}

	go func() {
		server.Serve(listener)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	ctx := context.Background()
	if p.ShutdownTimeout != time.Duration(0) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, p.ShutdownTimeout)
		defer cancel()
	}

	if err := server.Shutdown(ctx); err != nil {
		p.Logger.Fatal(err)
	}

	return nil
}
