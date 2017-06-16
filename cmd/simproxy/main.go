package main

import (
	"log"
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

func setupErrorLogger(c *ErrorLogConfig) (*log.Logger, error) {
	w, err := os.OpenFile(*c.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return log.New(w, "", log.LstdFlags), nil
}

func start(config *Config) {
	errorLogger, err := setupErrorLogger(config.ErrorLog)
	if err != nil {
		log.Fatal(err)
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
		f, err := os.OpenFile(*config.AccessLog.Path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			errorLogger.Fatal(err)
		}
		defer f.Close()

		accessLogger, err = simproxy.NewAccessLogger(*config.AccessLog.Format, f, config.AccessLog.Fields)
		if err != nil {
			errorLogger.Fatal(err)
		}
	}

	proxy := simproxy.NewProxy(balancer, accessLogger, *config.ReadTimeout, *config.WriteTimeout, errorLogger)
	err = proxy.ListenAndServe(*config.Listen)
	if err != nil {
		errorLogger.Fatal(err)
	}
}
