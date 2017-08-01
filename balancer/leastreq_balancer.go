package balancer

import (
	"errors"
	"fmt"
	"sync"

	"github.com/emirpasic/gods/sets/treeset"
	"github.com/ryotarai/simproxy/types"
)

type leastreqState struct {
	requests      int
	backend       *types.Backend
	totalRequests int64
	id            int
}

func leastreqStateComparator(a, b interface{}) int {
	itemA := a.(*leastreqState)
	itemB := b.(*leastreqState)

	weightA := itemA.backend.Weight
	weightB := itemB.backend.Weight

	if itemA == itemB {
		return 0
	}

	delta := float64(itemA.requests)/weightA -
		float64(itemB.requests)/weightB
	if delta < 0.0 {
		return -1
	} else if delta > 0.0 {
		return 1
	}

	delta = float64(itemA.totalRequests)/weightA -
		float64(itemB.totalRequests)/weightB
	if delta < 0.0 {
		return -1
	} else if delta > 0.0 {
		return 1
	}

	delta = itemB.backend.Weight - itemA.backend.Weight
	if delta < 0.0 {
		return -1
	} else if delta > 0.0 {
		return 1
	}

	return itemA.id - itemB.id
}

type LeastreqBalancer struct {
	set            *treeset.Set
	stateByBackend map[*types.Backend]*leastreqState
	mutex          *sync.Mutex
	currentID      int
}

func NewLeastreqBalancer() *LeastreqBalancer {
	b := &LeastreqBalancer{
		mutex:          &sync.Mutex{},
		stateByBackend: map[*types.Backend]*leastreqState{},
	}
	b.set = treeset.NewWith(leastreqStateComparator)

	return b
}

// for debugging
func (b *LeastreqBalancer) printState() {
	b.set.Each(func(i int, v interface{}) {
		item := v.(*leastreqState)
		fmt.Printf("%d: %+v %+v\n", i, item, item.backend)
	})
}

func (b *LeastreqBalancer) PickBackend() (*types.Backend, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.set.Empty() {
		return nil, errors.New("no backend is available")
	}

	iter := b.set.Iterator()
	if !iter.First() {
		return nil, nil
	}
	item := iter.Value().(*leastreqState)
	b.set.Remove(item)
	item.requests++
	// This can cause overflow but it cannot happen practically
	// because 9223372036854775807/10000rps/60s/60m/24h/356d = 29986514year
	item.totalRequests++
	b.set.Add(item)

	return item.backend, nil
}

func (b *LeastreqBalancer) ReturnBackend(backend *types.Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	item, ok := b.stateByBackend[backend]
	if !ok {
		return // already removed
	}

	b.set.Remove(item)
	if item.requests > 0 {
		item.requests--
	}
	b.set.Add(item)
}

func (b *LeastreqBalancer) AddBackend(backend *types.Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	_, ok := b.stateByBackend[backend]
	if ok {
		return
	}

	item := &leastreqState{
		id:       b.currentID,
		requests: 0,
		backend:  backend,
	}
	b.currentID++

	b.set.Add(item)
	b.stateByBackend[backend] = item
}

func (b *LeastreqBalancer) RemoveBackend(backend *types.Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	item, ok := b.stateByBackend[backend]
	if !ok {
		return
	}

	b.set.Remove(item)
	delete(b.stateByBackend, backend)
}
