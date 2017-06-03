package simproxy

import (
	"sync"

	"github.com/emirpasic/gods/sets/treeset"
)

type LeastreqState struct {
	Requests int
	Backend  *Backend
}

func leastreqStateComparator(a, b interface{}) int {
	itemA := a.(*LeastreqState)
	itemB := b.(*LeastreqState)

	if itemA == itemB {
		return 0
	}

	delta := float64(itemA.Requests)/float64(itemA.Backend.Weight) -
		float64(itemB.Requests)/float64(itemB.Backend.Weight)
	if delta < 0.0 {
		return -1
	}
	return 1
}

type LeastreqBalancer struct {
	set            *treeset.Set
	stateByBackend map[*Backend]*LeastreqState
	mutex          *sync.Mutex
}

func NewLeastreqBalancer(backends []*Backend) *LeastreqBalancer {
	b := &LeastreqBalancer{
		mutex:          &sync.Mutex{},
		stateByBackend: map[*Backend]*LeastreqState{},
	}
	b.set = treeset.NewWith(leastreqStateComparator)

	for _, backend := range backends {
		b.AddBackend(backend)
	}

	return b
}

func (b *LeastreqBalancer) PickBackend() *Backend {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	iter := b.set.Iterator()
	if !iter.First() {
		return nil
	}
	item := iter.Value().(*LeastreqState)
	b.set.Remove(item)
	item.Requests++
	b.set.Add(item)

	return item.Backend
}

func (b *LeastreqBalancer) ReturnBackend(backend *Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	item := b.stateByBackend[backend]
	b.set.Remove(item)
	if item.Requests > 0 {
		item.Requests--
	}
	b.set.Add(item)
}

func (b *LeastreqBalancer) AddBackend(backend *Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	item := &LeastreqState{
		Requests: 0,
		Backend:  backend,
	}
	b.set.Add(item)
	b.stateByBackend[backend] = item
}
