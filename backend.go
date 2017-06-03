package simproxy

import (
	"net/url"
)

type Backend struct {
	URL    *url.URL
	Weight int
}
