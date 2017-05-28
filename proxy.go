package simproxy

import (
	"net"
	"net/http"
	"net/http/httputil"
)

type Proxy struct {
	balancer Balancer
}

func NewProxy(balancer Balancer) *Proxy {
	return &Proxy{
		balancer: balancer,
	}
}

func (p *Proxy) Serve() error {
	transport := &Transport{
		Parent: http.DefaultTransport,
	}
	handler := &httputil.ReverseProxy{
		Director:  p.director,
		Transport: transport,
	}
	server := http.Server{
		Handler: handler,
	}

	l, err := p.listener()
	if err != nil {
		return err
	}

	defer l.Close()
	server.Serve(l)

	return nil
}

func (p *Proxy) listener() (net.Listener, error) {
	return net.Listen("tcp", "127.0.0.1:8080")
}

func (p *Proxy) director(req *http.Request) {
	backend := p.balancer.RetainServer()

	req.URL.Scheme = backend.URL.Scheme
	req.URL.Host = backend.URL.Host
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}
