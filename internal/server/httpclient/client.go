// Package httpclient initialises the http client with auth headers
package httpclient

import (
	"net/http"
)

type roundTripperWithHeaders struct {
	base    http.RoundTripper
	headers map[string]string
}

func (rt roundTripperWithHeaders) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range rt.headers {
		if v == "" {
			continue
		}
		req.Header.Set(k, v)
	}
	return rt.base.RoundTrip(req)
}

// Option configures the authenticated HTTP client/transport.
type Option interface {
	apply(*roundTripperWithHeaders)
}

type headersOption struct{ h map[string]string }

func (o headersOption) apply(r *roundTripperWithHeaders) { r.headers = o.h }

// WithHeaders sets default headers on all outgoing requests.
func WithHeaders(h map[string]string) Option { return headersOption{h: h} }

// NewHTTPClient builds an http.Client with options. Defaults: http.DefaultTransport, no headers.
func NewHTTPClient(opts ...Option) *http.Client {
	rt := roundTripperWithHeaders{base: http.DefaultTransport}
	for _, o := range opts {
		o.apply(&rt)
	}
	return &http.Client{Transport: rt}
}
