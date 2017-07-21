package simproxy

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealthStateFileStoreLoad(t *testing.T) {
	dir, err := ioutil.TempDir("", "simproxy")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "a.tsv")
	data := []byte("foo\t0\nbar\t1\n")
	ioutil.WriteFile(path, data, 0644)

	s := NewHealthStateFileStore(path)
	err = s.Load()
	if err != nil {
		t.Error(err)
	}

	if state := s.State("foo"); state != HEALTH_STATE_HEALTHY {
		t.Errorf("expected HEALTH_STATE_HEALTHY(0) but got %d", state)
	}
	if state := s.State("bar"); state != HEALTH_STATE_DEAD {
		t.Errorf("expected HEALTH_STATE_DEAD(1) but got %d", state)
	}
	if state := s.State("baz"); state != HEALTH_STATE_UNKNOWN {
		t.Errorf("expected HEALTH_STATE_UNKNOWN(2) but got %d", state)
	}
}

func TestHealthStateFileStoreMark(t *testing.T) {
	dir, err := ioutil.TempDir("", "simproxy")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "a.tsv")
	data := []byte("foo\t0\nbar\t1\n")
	ioutil.WriteFile(path, data, 0644)

	s := NewHealthStateFileStore(path)
	err = s.Load()
	if err != nil {
		t.Error(err)
	}

	s.Mark("foo", HEALTH_STATE_DEAD)
	s.Mark("bar", HEALTH_STATE_HEALTHY)

	b, err := ioutil.ReadFile(path)
	expected := "foo\t1\n"
	if !strings.Contains(string(b), expected) {
		t.Errorf("expected containing %+v but not: %+v", expected, string(b))
	}

	expected = "bar\t0\n"
	if !strings.Contains(string(b), expected) {
		t.Errorf("expected containing %+v but not: %+v", expected, string(b))
	}
}

func TestHealthStateFileStoreCleanup(t *testing.T) {
	dir, err := ioutil.TempDir("", "simproxy")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "a.tsv")
	data := []byte("foo\t0\nbar\t1\n")
	ioutil.WriteFile(path, data, 0644)

	s := NewHealthStateFileStore(path)
	err = s.Load()
	if err != nil {
		t.Error(err)
	}

	s.Cleanup([]string{"bar"})

	b, err := ioutil.ReadFile(path)
	expected := "bar\t1\n"
	if string(b) != expected {
		t.Errorf("expected %+v but got %+v", expected, string(b))
	}
}
