package balancer

import (
	"fmt"

	"github.com/ryotarai/simproxy/types"
)

type Balancer interface {
	AddBackend(*types.Backend)
	RemoveBackend(*types.Backend)
	PickBackend() (*types.Backend, error)
	ReturnBackend(*types.Backend)
}

func NewBalancer(method string) (Balancer, error) {
	switch method {
	case "leastreq":
		return NewLeastreqBalancer(), nil
	}
	return nil, fmt.Errorf("%s method is not vailid", method)
}
