package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"

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
		startPprof(*config.PprofAddr)
	}

	balancer, err := simproxy.NewBalancer(*config.BalancingMethod)
	if err != nil {
		errorLogger.Fatal(err)
	}

	healthStore := simproxy.NewHealthStateStore(*config.Healthcheck.StateFile)
	err = healthStore.Load()
	if err != nil {
		errorLogger.Fatal(err)
	}

	hcPath, err := url.Parse(*config.Healthcheck.Path)
	if err != nil {
		errorLogger.Fatal(err)
	}

	backendStrURLs := []string{}
	for _, b := range config.Backends {
		url, err := url.Parse(*b.URL)
		if err != nil {
			errorLogger.Fatal(err)
		}

		backendStrURLs = append(backendStrURLs, url.String())

		b2 := &simproxy.Backend{
			URL:            url,
			HealthcheckURL: url.ResolveReference(hcPath),
			Weight:         *b.Weight,
		}

		healthchecker := &simproxy.HealthChecker{
			State:     healthStore,
			Logger:    errorLogger,
			Backend:   b2,
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

	err = healthStore.Cleanup(backendStrURLs)
	if err != nil {
		errorLogger.Fatal(err)
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

	h := ""
	if config.BackendURLHeader != nil {
		h = *config.BackendURLHeader
	}

	maxIdleConnsPerHost := 0 // default
	if config.MaxIdleConnsPerHost != nil {
		maxIdleConnsPerHost = *config.MaxIdleConnsPerHost
	}

	maxIdleConns := 100
	if config.MaxIdleConns != nil {
		maxIdleConns = *config.MaxIdleConns
	}

	handler := simproxy.NewHandler(balancer, errorLogger, accessLogger, h, maxIdleConns, maxIdleConnsPerHost, config.EnableBackendTrace)

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
	err = proxy.ListenAndServe(*config.Listen)
	if err != nil {
		errorLogger.Fatal(err)
	}
}

func startPprof(addr string) {
	s := &http.Server{
		Addr:    addr,
		Handler: http.HandlerFunc(pprof.Index),
	}
	go func() {
		s.ListenAndServe()
	}()
}
