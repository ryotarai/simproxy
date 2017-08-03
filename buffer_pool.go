package simproxy

import (
	"sync"
)

type BufferPool struct {
	size    int
	buffers [][]byte
	mutex   sync.Mutex
}

func NewBufferPool(size int) *BufferPool {
	return &BufferPool{
		size:    size,
		buffers: [][]byte{},
		mutex:   sync.Mutex{},
	}
}

func (p *BufferPool) Get() []byte {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	l := len(p.buffers)
	if l > 0 {
		b := p.buffers[l-1]
		p.buffers = p.buffers[:l-1]
		return b
	}

	return make([]byte, p.size)
}

func (p *BufferPool) Put(b []byte) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.buffers = append(p.buffers, b)
}
