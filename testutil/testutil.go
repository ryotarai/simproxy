package testutil

// Utilities for testing

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/ryotarai/simproxy/types"
)

type testServer struct {
	Status int
	Server *httptest.Server
}

func NewTestServer() *testServer {
	s := &testServer{
		Status: 200,
	}
	s.Server = httptest.NewServer(http.HandlerFunc(s.handler))
	return s
}

func (s *testServer) URL() *url.URL {
	u, err := url.Parse(s.Server.URL)
	if err != nil {
		panic(err)
	}
	return u
}

func (s *testServer) handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("x-method", r.Method)
	w.Header().Add("x-url", r.URL.String())
	w.Header().Add("x-header", r.Header.Get("x-header"))
	w.Header().Add("x-remote-addr", r.RemoteAddr)
	w.WriteHeader(s.Status)

	body, _ := ioutil.ReadAll(r.Body)
	fmt.Fprintf(w, "%s", string(body))
}

type DummyBalancer struct {
	Backends []*types.Backend
}

func (b *DummyBalancer) AddBackend(be *types.Backend) {
	b.Backends = append(b.Backends, be)
}

func (b *DummyBalancer) RemoveBackend(b1 *types.Backend) {
	for i, b2 := range b.Backends {
		if b1 == b2 {
			b.Backends[i] = b.Backends[len(b.Backends)-1]
			b.Backends = b.Backends[:len(b.Backends)-1]
			break
		}
	}
}

func (b *DummyBalancer) PickBackend() (*types.Backend, error) {
	return b.Backends[0], nil
}

func (b *DummyBalancer) ReturnBackend(*types.Backend) {
	// no impl
}
