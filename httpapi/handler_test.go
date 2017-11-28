package httpapi

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerNotFound(t *testing.T) {
	h := NewHandler()
	s := httptest.NewServer(h)
	defer s.Close()

	r, err := http.Get(s.URL)
	assert.Nil(t, err)
	assert.Equal(t, 404, r.StatusCode)
}

func TestHandlerMetrics(t *testing.T) {
	h := NewHandler()
	s := httptest.NewServer(h)
	defer s.Close()

	r, err := http.Get(s.URL + "/metrics")
	assert.Nil(t, err)
	assert.Equal(t, 200, r.StatusCode)

	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(t, err)
	assert.Equal(t, "/metrics\n", string(body))
}
