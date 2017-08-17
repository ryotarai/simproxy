package handler

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/ryotarai/simproxy/testutil"
	"github.com/ryotarai/simproxy/types"
)

func TestCopyBuffer(t *testing.T) {
	p := ReverseProxy{}
	dirtyBuf := []byte{2, 2, 2, 2, 2, 2, 2, 2}
	srcBytes := []byte{1, 1, 1, 1}

	src := bytes.NewReader(srcBytes)
	dst := bytes.NewBuffer([]byte{})
	p.copyBuffer(dst, src, dirtyBuf)

	buf := make([]byte, 8)
	n, err := dst.Read(buf)
	if err != nil {
		t.Error(err)
	}
	if n != 4 {
		t.Errorf("length of written bytes is expected %d but got %d", 4, n)
	}
	expected := []byte{1, 1, 1, 1, 0, 0, 0, 0}
	for i, b := range expected {
		if buf[i] != b {
			t.Errorf("expected %d but got %d", b, buf[i])
		}
	}
}

func sendTestRequest(u *url.URL, req *http.Request, backendURLHeader string) *http.Response {
	backend := &types.Backend{
		URL: u,
	}
	balancer := &testutil.DummyBalancer{}
	balancer.AddBackend(backend)

	handler := &ReverseProxy{
		Balancer:         balancer,
		BackendURLHeader: backendURLHeader,
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	return w.Result()
}

func TestHandlerGET(t *testing.T) {
	ts := testutil.NewTestServer()
	defer ts.Server.Close()

	req := httptest.NewRequest("GET", "http://example.com/foo?a=b", nil)
	req.Header.Add("x-header", "hello")
	res := sendTestRequest(ts.URL(), req, "x-simproxy-backend")

	if res.Header.Get("x-method") != "GET" {
		t.Error("invalid method")
	}
	if res.Header.Get("x-url") != "/foo?a=b" {
		t.Error("invalid url")
	}
	if res.Header.Get("x-header") != "hello" {
		t.Error("invalid header")
	}
	if res.Header.Get("x-simproxy-backend") != ts.Server.URL {
		t.Error("invalid header")
	}
}

func TestHandlerKeepalive(t *testing.T) {
	ts := testutil.NewTestServer()
	defer ts.Server.Close()

	req1 := httptest.NewRequest("GET", "http://example.com/", nil)
	res1 := sendTestRequest(ts.URL(), req1, "")
	addr1 := res1.Header.Get("x-remote-addr")

	req2 := httptest.NewRequest("GET", "http://example.com/", nil)
	res2 := sendTestRequest(ts.URL(), req2, "")
	addr2 := res2.Header.Get("x-remote-addr")

	if addr1 != addr2 {
		t.Errorf("remote addrs are different: %s, %s", addr1, addr2)
	}
}

func TestHandlerPOST(t *testing.T) {
	ts := testutil.NewTestServer()
	defer ts.Server.Close()

	body := bytes.NewBufferString("THIS IS BODY")
	req := httptest.NewRequest("POST", "http://example.com/foo?a=b", body)
	req.Header.Add("x-header", "hello")
	res := sendTestRequest(ts.URL(), req, "")

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
