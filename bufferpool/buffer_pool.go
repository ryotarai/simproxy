package bufferpool

import (
	"sync"
)

type BufferPool struct {
	pool sync.Pool
}

func New(size int) *BufferPool {
	return &BufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return make([]byte, size)
			},
		},
	}
}

func (p *BufferPool) Get() []byte {
	b := p.pool.Get()
	return b.([]byte)
}

func (p *BufferPool) Put(b []byte) {
	p.pool.Put(b)
}
