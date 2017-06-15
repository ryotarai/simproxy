package simproxy

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

type HealthStateRepo struct {
	Path  string
	state map[string]HealthState
	mutex sync.Mutex
}

func NewHealthStateRepo(path string) *HealthStateRepo {
	return &HealthStateRepo{
		Path:  path,
		mutex: sync.Mutex{},
	}
}

func (s *HealthStateRepo) Load() error {
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

func (s *HealthStateRepo) Mark(u string, state HealthState) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.state[u] = state
	return s.write()
}

func (s *HealthStateRepo) write() error {
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
