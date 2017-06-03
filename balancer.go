package simproxy

type Balancer interface {
	PickBackend() *Backend
	ReturnBackend(*Backend)
}
}
