// Package httpclient initialises the http client with auth headers
package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

	resp, err := rt.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	if resp != nil && resp.Body != nil {
		bodyBytes, berr := io.ReadAll(resp.Body)
		if berr == nil {
			ct := resp.Header.Get("Content-Type")
			fmt.Printf("UPSTREAM RESP %s %s -> %d CT=%q\nBody: %s\n",
				req.Method, req.URL.String(), resp.StatusCode, ct, string(bodyBytes))

			// Fallback adapter: if upstream returns JSON (non-streamable),
			// wrap it as a single SSE "message" event so the SDK can read it.
			lct := strings.ToLower(ct)
			if strings.HasPrefix(lct, "application/json") && bytes.Contains(bodyBytes, []byte("\"jsonrpc\"")) {
				sse := []byte("event: message\n" + "data: " + string(bodyBytes) + "\n\n")
				resp.Header.Set("Content-Type", "text/event-stream")
				resp.Header.Set("Content-Length", strconv.Itoa(len(sse)))
				resp.ContentLength = int64(len(sse))
				resp.Body = io.NopCloser(bytes.NewBuffer(sse))
			} else {
				// restore body for downstream consumers untouched
				resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
	}

	return resp, nil
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
