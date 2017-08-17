package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func serveHTTPAndHandleSignal(server http.Server, listener net.Listener, timeout time.Duration) error {
	go func() {
		server.Serve(listener)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	ctx := context.Background()
	if timeout != time.Duration(0) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}
