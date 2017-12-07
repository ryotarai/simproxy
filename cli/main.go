package cli

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryotarai/simproxy/accesslogger"
	"github.com/ryotarai/simproxy/balancer"
	"github.com/ryotarai/simproxy/bufferpool"
	"github.com/ryotarai/simproxy/handler"
	"github.com/ryotarai/simproxy/health"
	"github.com/ryotarai/simproxy/httpapi"
	"github.com/ryotarai/simproxy/listener"
	"github.com/sirupsen/logrus"
)

func Start(args []string) {
	options := CommandLineOptions{}
	fs := setupFlagSet(args[0], &options)
	err := fs.Parse(args[1:])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
		fmt.Println(err)
		os.Exit(1)
	}

	err = config.Validate()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	start(config)
}

func openWritableFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
}

func setupLogger(path string, level *string) *logrus.Logger {
	w, err := openWritableFile(path)
	if err != nil {
		panic(err)
	}

	logger := logrus.New()
	logger.Out = w

	l := logrus.InfoLevel // default
	if level != nil {
		l, err = logrus.ParseLevel(*level)
		if err != nil {
			l = logrus.InfoLevel
			logger.WithField("err", err).Warn("Setting log level to INFO")
		}
	}
	logger.Level = l

	return logger
}

func start(config *Config) {
	logger := setupLogger(*config.Log.Path, config.Log.Level)
	logger.Infof("Starting Simproxy v%s", Version)

	balancer, err := balancer.NewBalancer(*config.BalancingMethod)
	if err != nil {
		logger.Fatal(err)
	}

	healthStore := health.NewHealthStateFileStore(*config.Healthcheck.StateFile)
	err = healthStore.Load()
	if err != nil {
		logger.Fatal(err)
	}

	backends, err := config.BuildBackends()
	if err != nil {
		logger.Fatal(err)
	}

	backendStrURLs := []string{}
	for _, b := range backends {
		backendStrURLs = append(backendStrURLs, b.URL.String())
	}
	err = healthStore.Cleanup(backendStrURLs)
	if err != nil {
		logger.Fatal(err)
	}

	for _, b := range backends {
		healthchecker := &health.HealthChecker{
			State:     healthStore,
			Logger:    logger,
			Backend:   b,
			Balancer:  balancer,
			Interval:  *config.Healthcheck.Interval,
			FallCount: *config.Healthcheck.FallCount,
			RiseCount: *config.Healthcheck.RiseCount,
		}
		err = healthchecker.Start()
		if err != nil {
			logger.Fatal(err)
		}
	}

	var accessLogger handler.AccessLogger
	if config.AccessLog != nil {
		f, err := openWritableFile(*config.AccessLog.Path)
		if err != nil {
			logger.Fatal(err)
		}
		defer f.Close()

		accessLogger, err = accesslogger.New(*config.AccessLog.Format, f, config.AccessLog.Fields)
		if err != nil {
			logger.Fatal(err)
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
		ErrorLog:            log.New(logger.WriterLevel(logrus.ErrorLevel), "", 0),
		AccessLogger:        accessLogger,
		BackendURLHeader:    backendURLHeader,
		EnableClientTrace:   config.EnableBackendTrace,
		Transport:           transport,
		AppendXForwardedFor: config.AppendXForwardedFor,
		BufferPool:          bufferPool,
	}

	server := http.Server{
		ErrorLog: log.New(logger.WriterLevel(logrus.ErrorLevel), "", 0),
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

	var shutdownTimeout time.Duration
	if config.ShutdownTimeout != nil {
		shutdownTimeout = *config.ShutdownTimeout
	}

	if a := config.HTTPAPIAddr; a != nil {
		logger.Infof("Enabling HTTP API on %s", *a)
		l, err := listener.Listen(*a)
		if err != nil {
			logger.Fatal(err)
		}
		httpapi.Start(l, balancer)
	}

	l, err := listener.Listen(*config.Listen)
	if err != nil {
		logger.Fatal(err)
	}
	defer l.Close()

	go func() {
		server.Serve(l)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)
	<-sigCh

	if d := config.ShutdownDelay; d != nil {
		logger.Infof("Waiting %s before shutting down...", *d)
		time.Sleep(*d)
	}

	logger.Info("Shutting down...")

	healthStore.Close()

	ctx := context.Background()
	if shutdownTimeout != time.Duration(0) {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()
	}

	if err := server.Shutdown(ctx); err != nil {
		logger.WithField("err", err).Error("Error during shutting down")
	}
}
