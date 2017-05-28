package main

import (
	"log"

	"net/url"

	"github.com/ryotarai/simproxy"
)

func main() {
	u, _ := url.Parse("http://google.com")
	backends := []*simproxy.Backend{
		{URL: u, Weight: 1},
	}
	balancer := simproxy.NewRoundrobinBalancer(backends)
	proxy := simproxy.NewProxy(balancer)
	err := proxy.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
