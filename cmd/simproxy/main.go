package main

import (
	"log"
	"net/url"

	"os"

	"github.com/ryotarai/simproxy"
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
	w, err := os.OpenFile(*c.Path, os.O_APPEND, 0644)
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

	hcPath, err := url.Parse(*config.Healthcheck.Path)
	if err != nil {
		errorLogger.Fatal(err)
	}

	for _, b := range config.Backends {
		url, err := url.Parse(*b.URL)
		if err != nil {
			errorLogger.Fatal(err)
		}

		b2 := &simproxy.Backend{
			URL:            url,
			HealthcheckURL: url.ResolveReference(hcPath),
			Weight:         *b.Weight,
		}

		healthchecker := &simproxy.Healthchecker{
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

	f, err := os.Open(*config.AccessLog.Path)
	if err != nil {
		errorLogger.Fatal(err)
	}
	defer f.Close()

	logger, err := simproxy.NewAccessLogger(*config.AccessLog.Format, f, config.AccessLog.Fields)
	if err != nil {
		errorLogger.Fatal(err)
	}

	proxy := simproxy.NewProxy(balancer, logger, *config.ReadTimeout, *config.WriteTimeout, errorLogger)
	err = proxy.ListenAndServe(*config.Listen)
	if err != nil {
		errorLogger.Fatal(err)
	}
}
