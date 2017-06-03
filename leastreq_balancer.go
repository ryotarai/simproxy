package simproxy

import (
	"sync"

	"github.com/emirpasic/gods/sets/treeset"
)

type SetItem struct {
	Requests int
	Backend  *Backend
}

type LeastreqBalancer struct {
	set           *treeset.Set
	itemByBackend map[*Backend]*SetItem
	mutex         *sync.Mutex
}

func NewLeastreqBalancer(backends []*Backend) *LeastreqBalancer {
	b := &LeastreqBalancer{
		mutex:         &sync.Mutex{},
		itemByBackend: map[*Backend]*SetItem{},
	}
	b.set = treeset.NewWith(b.setComparator)

	for _, backend := range backends {
		b.AddBackend(backend)
	}

	return b
}

func (b *LeastreqBalancer) RetainServer() *Backend {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	iter := b.set.Iterator()
	if !iter.First() {
		return nil
	}
	item := iter.Value().(*SetItem)
	b.set.Remove(item)
	item.Requests++
	b.set.Add(item)

	return item.Backend
}

func (b *LeastreqBalancer) ReleaseServer(backend *Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	item := b.itemByBackend[backend]
	b.set.Remove(item)
	if item.Requests > 0 {
		item.Requests--
	}
	b.set.Add(item)
}

func (b *LeastreqBalancer) AddBackend(backend *Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	item := &SetItem{
		Requests: 0,
		Backend:  backend,
	}
	b.set.Add(item)
	b.itemByBackend[backend] = item
}

func (balancer *LeastreqBalancer) setComparator(a, b interface{}) int {
	itemA := a.(*SetItem)
	itemB := b.(*SetItem)

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
