package simproxy

import (
	"sync"
)

type RoundrobinBalancer struct {
	mutex    sync.Mutex
	counter  uint64
	Backends []*Backend
}

func NewRoundrobinBalancer(backends []*Backend) *RoundrobinBalancer {
	return &RoundrobinBalancer{
		mutex:    sync.Mutex{},
		Backends: backends,
	}
}

func (b *RoundrobinBalancer) RetainServer() *Backend {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	backend := b.Backends[b.counter%uint64(len(b.Backends))]
	b.counter++
	return backend
}
