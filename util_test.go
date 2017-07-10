package simproxy

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type testServer struct {
	status int
	server *httptest.Server
}

func newTestServer() *testServer {
	s := &testServer{
		status: 200,
	}
	s.server = httptest.NewServer(http.HandlerFunc(s.handler))
	return s
}

func (s *testServer) url() *url.URL {
	u, err := url.Parse(s.server.URL)
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
	w.WriteHeader(s.status)

	body, _ := ioutil.ReadAll(r.Body)
	fmt.Fprintf(w, "%s", string(body))
}

type dummyBalancer struct {
	backends []*Backend
}

func (b *dummyBalancer) AddBackend(be *Backend) {
	b.backends = append(b.backends, be)
}

func (b *dummyBalancer) RemoveBackend(b1 *Backend) {
	for i, b2 := range b.backends {
		if b1 == b2 {
			b.backends[i] = b.backends[len(b.backends)-1]
			b.backends = b.backends[:len(b.backends)-1]
			break
		}
	}
}

func (b *dummyBalancer) PickBackend() (*Backend, error) {
	return b.backends[0], nil
}

func (b *dummyBalancer) ReturnBackend(*Backend) {
	// no impl
}
