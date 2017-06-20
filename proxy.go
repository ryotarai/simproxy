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
)

type Proxy struct {
	Logger       *log.Logger
	Handler      *Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewProxy(handler *Handler, logger *log.Logger) *Proxy {
	return &Proxy{
		Logger:  logger,
		Handler: handler,
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
			for _, l := range ls {
				l.Close()
			}
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
	server := http.Server{
		ErrorLog:     p.Logger,
		Handler:      p.Handler,
		ReadTimeout:  p.ReadTimeout,
		WriteTimeout: p.WriteTimeout,
	}

	go func() {
		server.Serve(listener)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), server.ReadTimeout+server.WriteTimeout+(1*time.Second))
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		p.Logger.Fatal(err)
	}

	return nil
}
