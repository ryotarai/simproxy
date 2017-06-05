package simproxy

import (
	"fmt"
)

type Balancer interface {
	AddBackend(*Backend)
	RemoveBackend(*Backend)
	PickBackend() (*Backend, error)
	ReturnBackend(*Backend)
}

func NewBalancer(method string) (Balancer, error) {
	switch method {
	case "leastreq":
		return NewLeastreqBalancer(), nil
	}
	return nil, fmt.Errorf("%s method is not vailid", method)
}
