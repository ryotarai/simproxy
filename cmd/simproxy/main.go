package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ryotarai/simproxy"
	"github.com/ryotarai/simproxy/handler"
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

	if config.PprofAddr != nil {
		startPprofServer(*config.PprofAddr)
	}

	balancer, err := simproxy.NewBalancer(*config.BalancingMethod)
	if err != nil {
		errorLogger.Fatal(err)
	}

	healthStore := simproxy.NewHealthStateFileStore(*config.Healthcheck.StateFile)
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
		healthchecker := &simproxy.HealthChecker{
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

		accessLogger, err = simproxy.NewAccessLogger(*config.AccessLog.Format, f, config.AccessLog.Fields)
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

	handler := &simproxy.Handler{
		Balancer:           balancer,
		Logger:             errorLogger,
		AccessLogger:       accessLogger,
		BackendURLHeader:   backendURLHeader,
		EnableBackendTrace: config.EnableBackendTrace,
		Transport:          transport,
	}
	handler.Setup()

	proxy := simproxy.NewProxy(handler, errorLogger)
	if config.ReadTimeout != nil {
		proxy.ReadTimeout = *config.ReadTimeout
	}
	if config.ReadHeaderTimeout != nil {
		proxy.ReadHeaderTimeout = *config.ReadHeaderTimeout
	}
	if config.WriteTimeout != nil {
		proxy.WriteTimeout = *config.WriteTimeout
	}
	if config.ShutdownTimeout != nil {
		proxy.ShutdownTimeout = *config.ShutdownTimeout
	}
	err = proxy.ListenAndServe(*config.Listen)
	if err != nil {
		errorLogger.Fatal(err)
	}
}
