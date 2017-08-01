package simproxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ryotarai/simproxy/types"
)

func sendTestRequest(u *url.URL, req *http.Request, backendURLHeader string) *http.Response {
	backend := &types.Backend{
		URL: u,
	}
	balancer := &dummyBalancer{}
	balancer.AddBackend(backend)

	handler := &Handler{
		Balancer:         balancer,
		BackendURLHeader: backendURLHeader,
	}
	handler.Setup()

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	return w.Result()
}

func TestHandlerGET(t *testing.T) {
	ts := newTestServer()
	defer ts.server.Close()

	req := httptest.NewRequest("GET", "http://example.com/foo?a=b", nil)
	req.Header.Add("x-header", "hello")
	res := sendTestRequest(ts.url(), req, "x-simproxy-backend")

	if res.Header.Get("x-method") != "GET" {
		t.Error("invalid method")
	}
	if res.Header.Get("x-url") != "/foo?a=b" {
		t.Error("invalid url")
	}
	if res.Header.Get("x-header") != "hello" {
		t.Error("invalid header")
	}
	if res.Header.Get("x-simproxy-backend") != ts.server.URL {
		t.Error("invalid header")
	}
}

func TestHandlerKeepalive(t *testing.T) {
	ts := newTestServer()
	defer ts.server.Close()

	req1 := httptest.NewRequest("GET", "http://example.com/", nil)
	res1 := sendTestRequest(ts.url(), req1, "")
	addr1 := res1.Header.Get("x-remote-addr")

	req2 := httptest.NewRequest("GET", "http://example.com/", nil)
	res2 := sendTestRequest(ts.url(), req2, "")
	addr2 := res2.Header.Get("x-remote-addr")

	if addr1 != addr2 {
		t.Errorf("remote addrs are different: %s, %s", addr1, addr2)
	}
}

func TestHandlerPOST(t *testing.T) {
	ts := newTestServer()
	defer ts.server.Close()

	body := bytes.NewBufferString("THIS IS BODY")
	req := httptest.NewRequest("POST", "http://example.com/foo?a=b", body)
	req.Header.Add("x-header", "hello")
	res := sendTestRequest(ts.url(), req, "")

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
