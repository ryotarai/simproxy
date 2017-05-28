package simproxy

import "net/http"

type Transport struct {
	Parent http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	res, err := t.Parent.RoundTrip(req)
	return res, err
}
