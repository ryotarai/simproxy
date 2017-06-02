package main

import (
	"log"

	"net/url"

	"github.com/ryotarai/simproxy"
)

func main() {
	u1, _ := url.Parse("http://127.0.0.1:9000")
	u2, _ := url.Parse("http://127.0.0.1:9001")
	backends := []*simproxy.Backend{
		{URL: u1, Weight: 1},
		{URL: u2, Weight: 2},
	}
	balancer := simproxy.NewRoundrobinBalancer(backends)
	proxy := simproxy.NewProxy(balancer)
	err := proxy.Serve()
	if err != nil {
		log.Fatal(err)
	}
}
