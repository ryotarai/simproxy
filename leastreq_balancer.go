package simproxy

import (
	"fmt"
	"sync"

	"errors"

	"github.com/emirpasic/gods/sets/treeset"
)

type LeastreqState struct {
	Requests int
	Backend  *Backend

	totalRequests int64
	id            int
}

func leastreqStateComparator(a, b interface{}) int {
	itemA := a.(*LeastreqState)
	itemB := b.(*LeastreqState)

	if itemA == itemB {
		return 0
	}

	delta := float64(itemA.Requests)/float64(itemA.Backend.Weight) -
		float64(itemB.Requests)/float64(itemB.Backend.Weight)
	if delta != 0 {
		if delta < 0.0 {
			return -1
		}
		return 1
	}

	d := itemB.Backend.Weight - itemA.Backend.Weight
	if d != 0 {
		return d
	}

	e := itemA.totalRequests - itemB.totalRequests
	if e < 0 {
		return -1
	} else if e > 0 {
		return 1
	}

	return itemA.id - itemB.id
}

type LeastreqBalancer struct {
	set            *treeset.Set
	stateByBackend map[*Backend]*LeastreqState
	mutex          *sync.Mutex
	currentID      int
}

func NewLeastreqBalancer() *LeastreqBalancer {
	b := &LeastreqBalancer{
		mutex:          &sync.Mutex{},
		stateByBackend: map[*Backend]*LeastreqState{},
	}
	b.set = treeset.NewWith(leastreqStateComparator)

	return b
}

// for debugging
func (b *LeastreqBalancer) printState() {
	b.set.Each(func(i int, v interface{}) {
		item := v.(*LeastreqState)
		fmt.Printf("%d: %+v %+v\n", i, item, item.Backend)
	})
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
	// This can cause overflow but it cannot happen practically
	// because 9223372036854775807/10000rps/60s/60m/24h/356d = 29986514year
	item.totalRequests++
	b.set.Add(item)

	return item.Backend, nil
}

func (b *LeastreqBalancer) ReturnBackend(backend *Backend) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	item, ok := b.stateByBackend[backend]
	if !ok {
		return // already removed
	}

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
		id:       b.currentID,
		Requests: 0,
		Backend:  backend,
	}
	b.currentID++

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
