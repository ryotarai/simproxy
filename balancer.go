package simproxy

import (
	"fmt"
)

type Balancer interface {
	PickBackend() *Backend
	ReturnBackend(*Backend)
}

func NewBalancer(method string, backends []*Backend) (Balancer, error) {
	switch method {
	case "roundrobin":
		return NewRoundrobinBalancer(backends), nil
	case "leastreq":
		return NewLeastreqBalancer(backends), nil
	}
	return nil, fmt.Errorf("%s method is not vailid", method)
}
