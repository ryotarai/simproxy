package simproxy

type LeastreqBalancer struct {
	Backends []*Backend
}

func NewLeastreqBalancer(backends []*Backend) *LeastreqBalancer {
	return &LeastreqBalancer{
		Backends: backends,
	}
}

func (b *LeastreqBalancer) RetainServer() *Backend {
}

func (b *LeastreqBalancer) ReleaseServer(*Backend) {
}
