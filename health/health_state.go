package health

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"sync"
)

type HealthState int

const (
	HEALTH_STATE_HEALTHY HealthState = iota
	HEALTH_STATE_DEAD
	HEALTH_STATE_UNKNOWN
)

type HealthStateStore interface {
	Load() error
	Mark(string, HealthState) error
	Cleanup([]string) error
	State(string) HealthState
}

type HealthStateFileStore struct {
	Path  string
	state map[string]HealthState
	mutex sync.Mutex
}

func NewHealthStateFileStore(path string) *HealthStateFileStore {
	return &HealthStateFileStore{
		Path:  path,
		mutex: sync.Mutex{},
	}
}

func (s *HealthStateFileStore) Load() error {
	state := map[string]HealthState{}

	if _, err := os.Stat(s.Path); os.IsNotExist(err) {
		s.state = state
		return nil
	}

	f, err := os.Open(s.Path)
	if err != nil {
		return err
	}

	scan := bufio.NewScanner(f)
	for scan.Scan() {
		fields := strings.Split(scan.Text(), "\t")
		i, err := strconv.Atoi(fields[1])
		if err != nil {
			return err
		}
		state[fields[0]] = HealthState(i)
	}

	s.state = state
	return nil
}

func (s *HealthStateFileStore) Mark(u string, state HealthState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state[u] = state
	return s.write()
}

func (s *HealthStateFileStore) Cleanup(us []string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	target := []string{}
	for u1 := range s.state {
		found := false
		for _, u2 := range us {
			if u1 == u2 {
				found = true
				break
			}
		}
		if !found {
			target = append(target, u1)
		}
	}

	for _, u := range target {
		delete(s.state, u)
	}
	return s.write()
}

func (s *HealthStateFileStore) State(u string) HealthState {
	st, ok := s.state[u]
	if ok {
		return st
	}
	return HEALTH_STATE_UNKNOWN
}

func (s *HealthStateFileStore) write() error {
	f, err := ioutil.TempFile("", "simproxy")
	if err != nil {
		return err
	}
	defer f.Close()

	for a, b := range s.state {
		f.WriteString(fmt.Sprintf("%s\t%d\n", a, b))
	}
	f.Close()

	err = os.Rename(f.Name(), s.Path)
	if err != nil {
		return err
	}

	return nil
}
