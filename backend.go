package simproxy

import (
	"net/url"
)

type Backend struct {
	URL            *url.URL
	HealthcheckURL *url.URL
	Weight         int
}

func (b *Backend) GetURL() *url.URL {
	return b.URL
}
