package simproxy

import (
	"net/url"
	"testing"
)

func TestRetainServer(t *testing.T) {
	u1, _ := url.Parse("http://127.0.0.1:9000")
	u2, _ := url.Parse("http://127.0.0.1:9001")
	backends := []*Backend{
		{URL: u1, Weight: 1},
		{URL: u2, Weight: 2},
	}
	b := NewLeastreqBalancer(backends)

	expects := []*Backend{
		backends[0],
		backends[1],
		backends[1],
		backends[0],
	}
	for _, e := range expects {
		if s := b.RetainServer(); s != e {
			t.Errorf("%#v is expected but %#v", e, s)
		}
	}
}

func TestReleaseServer(t *testing.T) {
	u1, _ := url.Parse("http://127.0.0.1:9000")
	u2, _ := url.Parse("http://127.0.0.1:9001")
	backends := []*Backend{
		{URL: u1, Weight: 1},
		{URL: u2, Weight: 2},
	}
	b := NewLeastreqBalancer(backends)

	expects := []*Backend{
		backends[0],
		backends[1],
		backends[1],
	}
	for _, e := range expects {
		if s := b.RetainServer(); s != e {
			t.Errorf("%#v is expected but %#v", e, s)
		}
	}

	b.ReleaseServer(backends[1])

	expects = []*Backend{
		backends[1],
		backends[0],
	}
	for _, e := range expects {
		if s := b.RetainServer(); s != e {
			t.Errorf("%#v is expected but %#v", e, s)
		}
	}
}
