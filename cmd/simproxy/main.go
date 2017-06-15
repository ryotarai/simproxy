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

	balancer, err := simproxy.NewBalancer(*config.BalancingMethod)
	if err != nil {
		log.Fatal(err)
	}

	hcPath, err := url.Parse(*config.Healthcheck.Path)
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range config.Backends {
		url, err := url.Parse(*b.URL)
		if err != nil {
			log.Fatal(err)
		}

		c := &simproxy.Backend{
			URL:            url,
			HealthcheckURL: url.ResolveReference(hcPath),
			Weight:         *b.Weight,
		}

		if err != nil {
			log.Fatal(err)
		}
		healthchecker := &simproxy.Healthchecker{
			Backend:   c,
			Balancer:  balancer,
			Interval:  *config.Healthcheck.Interval,
			FallCount: *config.Healthcheck.FallCount,
			RiseCount: *config.Healthcheck.RiseCount,
		}
		err = healthchecker.Start()
		if err != nil {
			log.Fatal(err)
		}
	}

	f, err := os.Open(*config.AccessLog.Path)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	logger, err := simproxy.NewAccessLogger(*config.AccessLog.Format, f, config.AccessLog.Fields)
	if err != nil {
		log.Fatal(err)
	}

	proxy := simproxy.NewProxy(balancer, logger)
	err = proxy.ListenAndServe(*config.Listen)
	if err != nil {
		log.Fatal(err)
	}
}
