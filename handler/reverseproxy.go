// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// This is forked version of https://github.com/golang/go/blob/2ea7d3461bb41d0ae12b56ee52d43314bcdb97f9/src/net/http/httputil/reverseproxy.go

// HTTP reverse proxy handler

package handler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/ryotarai/simproxy/types"
)

// onExitFlushLoop is a callback set by tests to detect the state of the
// flushLoop() goroutine.
var onExitFlushLoop func()

type LogRecord map[string]string

type AccessLogger interface {
	Log(LogRecord) error
}

type Backend interface {
	GetURL() *url.URL
}

type balancer interface {
	PickBackend() (*types.Backend, error)
	ReturnBackend(*types.Backend)
}

// ReverseProxy is an HTTP Handler that takes an incoming request and
// sends it to another server, proxying the response back to the
// client.
type ReverseProxy struct {
	AccessLogger AccessLogger

	Balancer            balancer
	BackendURLHeader    string
	EnableClientTrace   bool
	AppendXForwardedFor bool

	// The transport used to perform proxy requests.
	// If nil, http.DefaultTransport is used.
	Transport http.RoundTripper

	// FlushInterval specifies the flush interval
	// to flush to the client while copying the
	// response body.
	// If zero, no periodic flushing is done.
	FlushInterval time.Duration

	// ErrorLog specifies an optional logger for errors
	// that occur when attempting to proxy the request.
	// If nil, logging goes to os.Stderr via the log package's
	// standard logger.
	ErrorLog *log.Logger

	// BufferPool optionally specifies a buffer pool to
	// get byte slices for use by io.CopyBuffer when
	// copying HTTP response bodies.
	BufferPool BufferPool

	// ModifyResponse is an optional function that
	// modifies the Response from the backend.
	// If it returns an error, the proxy returns a StatusBadGateway error.
	ModifyResponse func(*http.Response) error
}

// A BufferPool is an interface for getting and returning temporary
// byte slices for use by io.CopyBuffer.
type BufferPool interface {
	Get() []byte
	Put([]byte)
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; http://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

func (p *ReverseProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	record := LogRecord{}
	logRW := wrapLogResponseWriter(rw)

	start := time.Now()
	p.serveHTTP(logRW, req, record)
	end := time.Now()

	if p.AccessLogger != nil {
		record["time"] = end.Format(time.RFC3339)
		record["time_nsec"] = fmt.Sprintf("%d", end.UnixNano())
		record["reqtime_nsec"] = fmt.Sprintf("%d", end.UnixNano()-start.UnixNano())
		record["status"] = fmt.Sprintf("%d", logRW.Status())
		record["size"] = fmt.Sprintf("%d", logRW.Size())
		record["path"] = req.URL.Path
		record["query"] = req.URL.RawQuery
		record["method"] = req.Method
		p.AccessLogger.Log(record)
	}
}

func (p *ReverseProxy) serveHTTP(rw http.ResponseWriter, req *http.Request, logRecord LogRecord) {
	transport := p.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	ctx := req.Context()
	if cn, ok := rw.(http.CloseNotifier); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		notifyChan := cn.CloseNotify()
		go func() {
			select {
			case <-notifyChan:
				cancel()
			case <-ctx.Done():
			}
		}()
	}

	outreq := req.WithContext(ctx) // includes shallow copies of maps, but okay
	if req.ContentLength == 0 {
		outreq.Body = nil // Issue 16036: nil Body for http.Transport retries
	}

	backend, err := p.Balancer.PickBackend()
	if err != nil {
		p.logf("http: proxy error: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	if p.EnableClientTrace {
		trace := p.clientTrace()
		outreq = outreq.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	}

	outreq.Header = cloneHeader(req.Header)

	p.director(outreq, backend.GetURL())
	logRecord["backend"] = backend.GetURL().String()

	outreq.Close = false

	// Remove hop-by-hop headers listed in the "Connection" header.
	// See RFC 2616, section 14.10.
	if c := outreq.Header.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				outreq.Header.Del(f)
			}
		}
	}

	// Remove hop-by-hop headers to the backend. Especially
	// important is "Connection" because we want a persistent
	// connection, regardless of what the client sent to us.
	for _, h := range hopHeaders {
		if outreq.Header.Get(h) != "" {
			outreq.Header.Del(h)
		}
	}

	if p.AppendXForwardedFor {
		if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			// If we aren't the first proxy retain prior
			// X-Forwarded-For information as a comma+space
			// separated list and fold multiple headers into one.
			if prior, ok := outreq.Header["X-Forwarded-For"]; ok {
				clientIP = strings.Join(prior, ", ") + ", " + clientIP
			}
			outreq.Header.Set("X-Forwarded-For", clientIP)
		}
	}

	res, err := transport.RoundTrip(outreq)
	p.Balancer.ReturnBackend(backend)

	if err != nil {
		p.logf("http: proxy error: %v", err)
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	// Remove hop-by-hop headers listed in the
	// "Connection" header of the response.
	if c := res.Header.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			if f = strings.TrimSpace(f); f != "" {
				res.Header.Del(f)
			}
		}
	}

	for _, h := range hopHeaders {
		res.Header.Del(h)
	}

	if p.ModifyResponse != nil {
		if err := p.ModifyResponse(res); err != nil {
			p.logf("http: proxy error: %v", err)
			rw.WriteHeader(http.StatusBadGateway)
			return
		}
	}

	if p.BackendURLHeader != "" {
		rw.Header().Add(p.BackendURLHeader, backend.GetURL().String())
	}

	copyHeader(rw.Header(), res.Header)

	// The "Trailer" header isn't included in the Transport's response,
	// at least for *http.Transport. Build it up from Trailer.
	announcedTrailers := len(res.Trailer)
	if announcedTrailers > 0 {
		trailerKeys := make([]string, 0, len(res.Trailer))
		for k := range res.Trailer {
			trailerKeys = append(trailerKeys, k)
		}
		rw.Header().Add("Trailer", strings.Join(trailerKeys, ", "))
	}

	rw.WriteHeader(res.StatusCode)
	if len(res.Trailer) > 0 {
		// Force chunking if we saw a response trailer.
		// This prevents net/http from calculating the length for short
		// bodies and adding a Content-Length.
		if fl, ok := rw.(http.Flusher); ok {
			fl.Flush()
		}
	}
	p.copyResponse(rw, res.Body)
	res.Body.Close() // close now, instead of defer, to populate res.Trailer

	if len(res.Trailer) == announcedTrailers {
		copyHeader(rw.Header(), res.Trailer)
		return
	}

	for k, vv := range res.Trailer {
		k = http.TrailerPrefix + k
		for _, v := range vv {
			rw.Header().Add(k, v)
		}
	}
}

func (p *ReverseProxy) clientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(dnsInfo httptrace.DNSStartInfo) {
			p.logf("TRACE: DNSStart: %+v", dnsInfo)
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			p.logf("TRACE: DNSDone: %+v", dnsInfo)
		},
		GetConn: func(hostPort string) {
			p.logf("TRACE: GetConn: %+v", hostPort)
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			p.logf("TRACE: GotConn: %+v", connInfo)
		},
		PutIdleConn: func(err error) {
			p.logf("TRACE: PutIdleConn: %+v", err)
		},
		ConnectStart: func(network, addr string) {
			p.logf("TRACE: ConnectStart: %+v %+v", network, addr)
		},
		ConnectDone: func(network, addr string, err error) {
			p.logf("TRACE: ConnectDone: %+v %+v %+v", network, addr, err)
		},
	}
}

func (p *ReverseProxy) copyResponse(dst io.Writer, src io.Reader) {
	if p.FlushInterval != 0 {
		if wf, ok := dst.(writeFlusher); ok {
			mlw := &maxLatencyWriter{
				dst:     wf,
				latency: p.FlushInterval,
				done:    make(chan bool),
			}
			go mlw.flushLoop()
			defer mlw.stop()
			dst = mlw
		}
	}

	var buf []byte
	if p.BufferPool != nil {
		buf = p.BufferPool.Get()
	}
	p.copyBuffer(dst, src, buf)
	if p.BufferPool != nil {
		p.BufferPool.Put(buf)
	}
}

func (p *ReverseProxy) copyBuffer(dst io.Writer, src io.Reader, buf []byte) (int64, error) {
	if len(buf) == 0 {
		buf = make([]byte, 32*1024)
	}
	var written int64
	for {
		nr, rerr := src.Read(buf)
		if rerr != nil && rerr != io.EOF && rerr != context.Canceled {
			p.logf("httputil: ReverseProxy read error during body copy: %v", rerr)
		}
		if nr > 0 {
			nw, werr := dst.Write(buf[:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if werr != nil {
				return written, werr
			}
			if nr != nw {
				return written, io.ErrShortWrite
			}
		}
		if rerr != nil {
			return written, rerr
		}
	}
}

func (p *ReverseProxy) logf(format string, args ...interface{}) {
	if p.ErrorLog != nil {
		p.ErrorLog.Printf(format, args...)
	} else {
		log.Printf(format, args...)
	}
}

func (p *ReverseProxy) director(req *http.Request, target *url.URL) {
	targetQuery := target.RawQuery

	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
	if targetQuery == "" || req.URL.RawQuery == "" {
		req.URL.RawQuery = targetQuery + req.URL.RawQuery
	} else {
		req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
	}
	if _, ok := req.Header["User-Agent"]; !ok {
		// explicitly disable User-Agent so it's not set to default value
		req.Header.Set("User-Agent", "")
	}
}

type writeFlusher interface {
	io.Writer
	http.Flusher
}

type maxLatencyWriter struct {
	dst     writeFlusher
	latency time.Duration

	mu   sync.Mutex // protects Write + Flush
	done chan bool
}

func (m *maxLatencyWriter) Write(p []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.dst.Write(p)
}

func (m *maxLatencyWriter) flushLoop() {
	t := time.NewTicker(m.latency)
	defer t.Stop()
	for {
		select {
		case <-m.done:
			if onExitFlushLoop != nil {
				onExitFlushLoop()
			}
			return
		case <-t.C:
			m.mu.Lock()
			m.dst.Flush()
			m.mu.Unlock()
		}
	}
}

func (m *maxLatencyWriter) stop() { m.done <- true }
