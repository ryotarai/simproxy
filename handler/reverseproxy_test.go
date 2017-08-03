package handler

import (
	"bytes"
	"testing"
)

func TestCopyBuffer(t *testing.T) {
	p := ReverseProxy{}
	dirtyBuf := []byte{2, 2, 2, 2, 2, 2, 2, 2}
	srcBytes := []byte{1, 1, 1, 1}

	src := bytes.NewReader(srcBytes)
	dst := bytes.NewBuffer([]byte{})
	p.copyBuffer(dst, src, dirtyBuf)

	buf := make([]byte, 8)
	n, err := dst.Read(buf)
	if err != nil {
		t.Error(err)
	}
	if n != 4 {
		t.Errorf("length of written bytes is expected %d but got %d", 4, n)
	}
	expected := []byte{1, 1, 1, 1, 0, 0, 0, 0}
	for i, b := range expected {
		if buf[i] != b {
			t.Errorf("expected %d but got %d", b, buf[i])
		}
	}
}
