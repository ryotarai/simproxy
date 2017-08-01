package types

import (
	"net/url"
)

type Backend struct {
	URL            *url.URL
	HealthcheckURL *url.URL
	Weight         float64
}

// GetURL satisfies handler.Backend interface
func (b *Backend) GetURL() *url.URL {
	return b.URL
}
