package httpapi

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ryotarai/simproxy/types"
	"github.com/stretchr/testify/assert"
)

type dummyBalancer struct {
}

func (b dummyBalancer) Metrics() map[*types.Backend]map[string]int64 {
	m := map[*types.Backend]map[string]int64{}
	u, err := url.Parse("http://example.com:8080/foo/")
	if err != nil {
		panic(err)
	}
	be := &types.Backend{
		URL:    u,
		Weight: 123,
	}
	m[be] = map[string]int64{
		"key1": 1,
		"key2": 2,
	}
	return m
}

func TestHandlerNotFound(t *testing.T) {
	h := NewHandler(dummyBalancer{})
	s := httptest.NewServer(h)
	defer s.Close()

	r, err := http.Get(s.URL)
	assert.Nil(t, err)
	assert.Equal(t, 404, r.StatusCode)
}

func TestHandlerMetrics(t *testing.T) {
	h := NewHandler(dummyBalancer{})
	s := httptest.NewServer(h)
	defer s.Close()

	r, err := http.Get(s.URL + "/metrics")
	assert.Nil(t, err)
	assert.Equal(t, 200, r.StatusCode)

	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err)
	assert.Equal(t, "simproxy_backend_weight{backend=http://example.com:8080/foo/} 123.000000\nsimproxy_key1{backend=http://example.com:8080/foo/} 1\nsimproxy_key2{backend=http://example.com:8080/foo/} 2\n", string(body))
}
