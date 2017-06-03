package main

import (
	"log"

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

	backends := []*simproxy.Backend{}
	for _, b := range config.Backends {
		c, err := b.Backend()
		if err != nil {
			log.Fatal(err)
		}
		backends = append(backends, c)
	}

	balancer, err := simproxy.NewBalancer(*config.BalancingMethod, backends)
	if err != nil {
		log.Fatal(err)
	}

	proxy := simproxy.NewProxy(balancer)
	err = proxy.Serve(*config.Listen)
	if err != nil {
		log.Fatal(err)
	}
}
