package simproxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type dummyBalancer struct {
	backend *Backend
}

func (b *dummyBalancer) AddBackend(*Backend) {
}

func (b *dummyBalancer) RemoveBackend(*Backend) {
}

func (b *dummyBalancer) PickBackend() (*Backend, error) {
	return b.backend, nil
}

func (b *dummyBalancer) ReturnBackend(*Backend) {
}

func setupTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("x-method", r.Method)
		w.Header().Add("x-url", r.URL.String())
		w.Header().Add("x-header", r.Header.Get("x-header"))
		w.Header().Add("x-remote-addr", r.RemoteAddr)

		body, _ := ioutil.ReadAll(r.Body)
		fmt.Fprintf(w, "%s", string(body))
	}))
}

func sendTestRequest(ts *httptest.Server, req *http.Request) *http.Response {
	u, err := url.Parse(ts.URL)
	if err != nil {
		panic(err)
	}

	backend := &Backend{
		URL: u,
	}
	balancer := &dummyBalancer{
		backend: backend,
	}
	handler := NewHandler(balancer, nil, nil)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	return w.Result()
}

func TestHandlerGET(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	req := httptest.NewRequest("GET", "http://example.com/foo?a=b", nil)
	req.Header.Add("x-header", "hello")
	res := sendTestRequest(ts, req)

	if res.Header.Get("x-method") != "GET" {
		t.Error("invalid method")
	}
	if res.Header.Get("x-url") != "/foo?a=b" {
		t.Error("invalid url")
	}
	if res.Header.Get("x-header") != "hello" {
		t.Error("invalid header")
	}
}

func TestHandlerKeepalive(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	req1 := httptest.NewRequest("GET", "http://example.com/", nil)
	res1 := sendTestRequest(ts, req1)
	addr1 := res1.Header.Get("x-remote-addr")

	req2 := httptest.NewRequest("GET", "http://example.com/", nil)
	res2 := sendTestRequest(ts, req2)
	addr2 := res2.Header.Get("x-remote-addr")

	if addr1 != addr2 {
		t.Errorf("remote addrs are different: %s, %s", addr1, addr2)
	}
}

func TestHandlerPOST(t *testing.T) {
	ts := setupTestServer()
	defer ts.Close()

	body := bytes.NewBufferString("THIS IS BODY")
	req := httptest.NewRequest("POST", "http://example.com/foo?a=b", body)
	req.Header.Add("x-header", "hello")
	res := sendTestRequest(ts, req)

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != "THIS IS BODY" {
		t.Error("invalid body")
	}
	if res.Header.Get("x-method") != "POST" {
		t.Error("invalid method")
	}
	if res.Header.Get("x-url") != "/foo?a=b" {
		t.Error("invalid url")
	}
	if res.Header.Get("x-header") != "hello" {
		t.Error("invalid header")
	}
}
