package simproxy

import (
	"net"
	"net/http"

	"github.com/ryotarai/simproxy/httputil"
)

type Proxy struct {
	balancer Balancer
}

func NewProxy(balancer Balancer) *Proxy {
	return &Proxy{
		balancer: balancer,
	}
}

func (p *Proxy) Serve(listen string) error {
	handler := &httputil.ReverseProxy{
		Director: p.director,
	}
	server := http.Server{
		Handler: handler,
	}

	l, err := net.Listen("tcp", listen)
	if err != nil {
		return err
	}
	defer l.Close()

	server.Serve(l)

	return nil
}

func (p *Proxy) director(req *http.Request) (func(), func()) {
	backend := p.balancer.PickBackend()
	target := backend.URL

	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = httputil.SingleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}

	afterRoundTrip := func() {
		p.balancer.ReturnBackend(backend)
	}

	return nil, afterRoundTrip
}
