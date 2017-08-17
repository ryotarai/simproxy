package simproxy

import (
	"testing"

	"github.com/ryotarai/simproxy/testutil"
	"github.com/ryotarai/simproxy/types"
)

type dummyHealthStateStore struct {
}

func (s dummyHealthStateStore) Load() error {
	return nil
}
func (s dummyHealthStateStore) Mark(string, HealthState) error {
	return nil
}
func (s dummyHealthStateStore) Cleanup([]string) error {
	return nil
}
func (s dummyHealthStateStore) State(string) HealthState {
	return HEALTH_STATE_UNKNOWN
}

func TestCheck(t *testing.T) {
	ts := testutil.NewTestServer()

	backend := &types.Backend{
		URL:            ts.URL(),
		HealthcheckURL: ts.URL(),
	}

	balancer := &testutil.DummyBalancer{}

	c := &HealthChecker{
		State:     dummyHealthStateStore{},
		Backend:   backend,
		Balancer:  balancer,
		FallCount: 2,
		RiseCount: 2,
		active:    false,
	}

	c.check()
	if len(balancer.Backends) != 0 {
		t.Error("expected that no backend is registered")
	}
	c.check()
	if len(balancer.Backends) != 1 || balancer.Backends[0] != backend {
		t.Error("expected that backend is registered")
	}
	ts.Status = 500
	c.check()
	if len(balancer.Backends) != 1 || balancer.Backends[0] != backend {
		t.Error("expected that backend is registered")
	}
	c.check()
	if len(balancer.Backends) != 0 {
		t.Error("expected that no backend is registered")
	}
}
