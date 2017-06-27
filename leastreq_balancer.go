package simproxy

import (
	"sync"

	"errors"

	"github.com/emirpasic/gods/sets/treeset"
)

type LeastreqState struct {
	Requests int
	Backend  *Backend
}

type LeastreqBalancer struct {
	set            *treeset.Set
	stateByBackend map[*Backend]*LeastreqState
	mutex          *sync.Mutex
}

func NewLeastreqBalancer() *LeastreqBalancer {
	b := &LeastreqBalancer{
		mutex:          &sync.Mutex{},
		stateByBackend: map[*Backend]*LeastreqState{},
	}
	b.set = treeset.NewWith(b.leastreqStateComparator)

	return b
}

func (bl *LeastreqBalancer) leastreqStateComparator(a, b interface{}) int {
	bl.mutex.Lock()
	defer bl.mutex.Unlock()

	itemA := a.(*LeastreqState)
	itemB := b.(*LeastreqState)

	if itemA == itemB {
		return 0
	}

	delta := float64(itemA.Requests)/float64(itemA.Backend.Weight) -
		float64(itemB.Requests)/float64(itemB.Backend.Weight)
	if delta == 0 {
		if itemA.Backend.URL.String() > itemB.Backend.URL.String() {
			return 1
		}
		return -1
	}
	if delta < 0.0 {
		return -1
	}
	return 1
}

func (b *LeastreqBalancer) PickBackend() (*Backend, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.set.Empty() {
		return nil, errors.New("no backend is available")
	}

	iter := b.set.Iterator()
	if !iter.First() {
		return nil, nil
	}
	item := iter.Value().(*LeastreqState)
	b.set.Remove(item)
	item.Requests++
	b.set.Add(item)

	return item.Backend, nil
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

	_, ok := b.stateByBackend[backend]
	if ok {
		return
	}

	item := &LeastreqState{
		Requests: 0,
		Backend:  backend,
	}
	b.set.Add(item)
	b.stateByBackend[backend] = item
}

func (b *LeastreqBalancer) RemoveBackend(backend *Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	item, ok := b.stateByBackend[backend]
	if !ok {
		return
	}

	b.set.Remove(item)
	delete(b.stateByBackend, backend)
}
