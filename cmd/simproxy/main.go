package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ryotarai/simproxy/accesslogger"
	"github.com/ryotarai/simproxy/balancer"
	"github.com/ryotarai/simproxy/bufferpool"
	"github.com/ryotarai/simproxy/handler"
	"github.com/ryotarai/simproxy/health"
	"github.com/ryotarai/simproxy/listener"
)

func main() {
	options := CommandLineOptions{}
	fs := setupFlagSet(os.Args[0], &options)
	err := fs.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if options.ShowVersion {
		fmt.Printf("simproxy v%s\n", Version)
		os.Exit(0)
	}

	if options.Config == "" {
		fmt.Println("ERROR: -config is mandatory")
		os.Exit(1)
	}

	config, err := LoadConfigFromYAML(options.Config)
	if err != nil {
		log.Fatal(err)
	}

	err = config.Validate()
	if err != nil {
		log.Fatal(err)
	}

	start(config)
}

func openWritableFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func start(config *Config) {
	w, err := openWritableFile(*config.ErrorLog.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()

	errorLogger := log.New(w, "", log.LstdFlags)
	errorLogger.Printf("INFO: Simproxy v%s", Version)

	if config.PprofAddr != nil {
		startPprofServer(*config.PprofAddr)
	}

	balancer, err := balancer.NewBalancer(*config.BalancingMethod)
	if err != nil {
		errorLogger.Fatal(err)
	}

	healthStore := health.NewHealthStateFileStore(*config.Healthcheck.StateFile)
	err = healthStore.Load()
	if err != nil {
		errorLogger.Fatal(err)
	}

	backends, err := config.BuildBackends()
	if err != nil {
		errorLogger.Fatal(err)
	}

	backendStrURLs := []string{}
	for _, b := range backends {
		backendStrURLs = append(backendStrURLs, b.URL.String())
	}
	err = healthStore.Cleanup(backendStrURLs)
	if err != nil {
		errorLogger.Fatal(err)
	}

	for _, b := range backends {
		healthchecker := &health.HealthChecker{
			State:     healthStore,
			Logger:    errorLogger,
			Backend:   b,
			Balancer:  balancer,
			Interval:  *config.Healthcheck.Interval,
			FallCount: *config.Healthcheck.FallCount,
			RiseCount: *config.Healthcheck.RiseCount,
		}
		err = healthchecker.Start()
		if err != nil {
			errorLogger.Fatal(err)
		}
	}

	var accessLogger handler.AccessLogger
	if config.AccessLog != nil {
		f, err := openWritableFile(*config.AccessLog.Path)
		if err != nil {
			errorLogger.Fatal(err)
		}
		defer f.Close()

		accessLogger, err = accesslogger.New(*config.AccessLog.Format, f, config.AccessLog.Fields)
		if err != nil {
			errorLogger.Fatal(err)
		}
	}

	backendURLHeader := ""
	if config.BackendURLHeader != nil {
		backendURLHeader = *config.BackendURLHeader
	}

	maxIdleConnsPerHost := 0 // DefaultMaxIdleConnsPerHost will be used
	if config.MaxIdleConnsPerHost != nil {
		maxIdleConnsPerHost = *config.MaxIdleConnsPerHost
	}

	maxIdleConns := 100
	if config.MaxIdleConns != nil {
		maxIdleConns = *config.MaxIdleConns
	}

	transport := &http.Transport{
		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConnsPerHost,

		// The following is the same as DefaultTransport
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	var bufferPool handler.BufferPool
	if config.EnableBufferPool {
		bufferPool = bufferpool.New(32 * 1024)
	}

	handler := &handler.ReverseProxy{
		Balancer:            balancer,
		ErrorLog:            errorLogger,
		AccessLogger:        accessLogger,
		BackendURLHeader:    backendURLHeader,
		EnableClientTrace:   config.EnableBackendTrace,
		Transport:           transport,
		AppendXForwardedFor: config.AppendXForwardedFor,
		BufferPool:          bufferPool,
	}

	server := http.Server{
		ErrorLog: errorLogger,
		Handler:  handler,
	}
	if config.ReadTimeout != nil {
		server.ReadTimeout = *config.ReadTimeout
	}
	if config.ReadHeaderTimeout != nil {
		server.ReadHeaderTimeout = *config.ReadHeaderTimeout
	}
	if config.WriteTimeout != nil {
		server.WriteTimeout = *config.WriteTimeout
	}

	listener, err := listener.Listen(*config.Listen)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer listener.Close()

	var shutdownTimeout time.Duration
	if config.ShutdownTimeout != nil {
		shutdownTimeout = *config.ShutdownTimeout
	}

	err = serveHTTPAndHandleSignal(server, listener, shutdownTimeout)
	if err != nil {
		errorLogger.Fatal(err)
	}
}
