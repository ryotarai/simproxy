package simproxy

import (
	"fmt"
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
	b := NewLeastreqBalancer()
	for _, be := range backends {
		b.AddBackend(be)
	}

	expects := []*Backend{
		backends[0],
		backends[1],
		backends[1],
		backends[0],
	}
	for _, e := range expects {
		s, err := b.PickBackend()
		if err != nil {
			t.Error(err)
		}

		if s != e {
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
	b := NewLeastreqBalancer()
	for _, be := range backends {
		b.AddBackend(be)
	}

	expects := []*Backend{
		backends[0],
		backends[1],
		backends[1],
	}
	for _, e := range expects {
		s, err := b.PickBackend()
		if err != nil {
			t.Error(err)
		}

		if s != e {
			t.Errorf("%#v is expected but %#v", e, s)
		}
	}

	b.ReturnBackend(backends[1])

	expects = []*Backend{
		backends[1],
		backends[0],
	}
	for _, e := range expects {
		s, err := b.PickBackend()
		if err != nil {
			t.Error(err)
		}

		if s != e {
			t.Errorf("%#v is expected but %#v", e, s)
		}
	}
}

func BenchmarkPickBackend1000(b *testing.B) {
	balancer := NewLeastreqBalancer()
	for i := 0; i < 1000; i++ {
		u, err := url.Parse(fmt.Sprintf("http://127.0.0.1/%d", i))
		if err != nil {
			b.Error(err)
		}
		balancer.AddBackend(&Backend{
			URL:    u,
			Weight: 1,
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		balancer.PickBackend()
	}
}

func BenchmarkReturnBackend1000(b *testing.B) {
	balancer := NewLeastreqBalancer()
	for i := 0; i < 1000; i++ {
		u, err := url.Parse(fmt.Sprintf("http://127.0.0.1/%d", i))
		if err != nil {
			b.Error(err)
		}
		balancer.AddBackend(&Backend{
			URL:    u,
			Weight: 1,
		})
	}

	bes := []*Backend{}
	for i := 0; i < b.N; i++ {
		be, err := balancer.PickBackend()
		if err != nil {
			b.Error(err)
		}
		bes = append(bes, be)
	}

	b.ResetTimer()

	for _, be := range bes {
		balancer.ReturnBackend(be)
	}
}
