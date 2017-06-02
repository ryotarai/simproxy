package simproxy

type Balancer interface {
	RetainServer() *Backend
	ReleaseServer(*Backend)
}
