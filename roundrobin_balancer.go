package simproxy

import (
	"sync"
)

type RoundrobinBalancer struct {
	mutex       sync.Mutex
	counter     uint64
	totalWeight uint64
	backendMap  map[uint64]*Backend
	Backends    []*Backend
}

func NewRoundrobinBalancer(backends []*Backend) *RoundrobinBalancer {
	totalWeight := uint64(0)
	for _, b := range backends {
		totalWeight += b.Weight
	}

	backendMap := map[uint64]*Backend{}
	i := uint64(0)
	for _, b := range backends {
		for j := uint64(0); j < b.Weight; j++ {
			backendMap[i] = b
			i++
		}
	}

	return &RoundrobinBalancer{
		mutex:       sync.Mutex{},
		totalWeight: totalWeight,
		backendMap:  backendMap,
		Backends:    backends,
	}
}

func (b *RoundrobinBalancer) RetainServer() *Backend {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	backend := b.backendMap[b.counter]
	b.counter++
	if b.counter >= b.totalWeight {
		b.counter = 0
	}
	return backend
}

func (b *RoundrobinBalancer) ReleaseServer(*Backend) {
}
