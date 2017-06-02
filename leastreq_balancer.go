package simproxy

import (
	"github.com/emirpasic/gods/sets/treeset"
)

type SetItem struct {
	Requests int
	Backend  *Backend
}

type LeastreqBalancer struct {
	set *treeset.Set
}

func NewLeastreqBalancer(backends []*Backend) *LeastreqBalancer {
	b := &LeastreqBalancer{}
	b.set = treeset.NewWith(b.setComparator)

	for _, backend := range backends {
		b.AddBackend(backend)
	}

	return b
}

func (b *LeastreqBalancer) RetainServer() *Backend {
	return nil
}

func (b *LeastreqBalancer) ReleaseServer(*Backend) {
}

func (b *LeastreqBalancer) AddBackend(backend *Backend) {
	item := &SetItem{
		Requests: 0,
		Backend:  backend,
	}
	b.set.Add(item)
}

func (balancer *LeastreqBalancer) setComparator(a, b interface{}) int {
	itemA := a.(*SetItem)
	itemB := b.(*SetItem)

	delta := itemA.Requests - itemB.Requests
	if delta != 0 {
		return delta
	}

	return itemA.Backend - itemB.Backend
}
