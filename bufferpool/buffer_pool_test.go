package bufferpool

import (
	"fmt"
	"testing"
)

func TestBufferPool(t *testing.T) {
	p := New(1024)
	b1 := p.Get()
	if len(b1) != 1024 {
		t.Errorf("unexpected buffer size %d", len(b1))
	}
	b2 := p.Get()
	if fmt.Sprintf("%p", b1) == fmt.Sprintf("%p", b2) {
		t.Error("the same buffer is returned")
	}
	p.Put(b1)
	b3 := p.Get()
	if fmt.Sprintf("%p", b1) != fmt.Sprintf("%p", b3) {
		t.Error("the different buffer is returned")
	}
}
