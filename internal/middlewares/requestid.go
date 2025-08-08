// Package middlewares contains HTTP middlewares used across the gateway.
package middlewares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"

	ck "github.com/ChiragChiranjib/mcp-proxy/internal/contextkey"
)

// RequestID injects an X-Request-Id header and stores it in context for tracing.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get("X-Request-Id")
			if rid == "" {
				rid = newID()
			}
			w.Header().Set("X-Request-Id", rid)
			ctx := context.WithValue(r.Context(), ck.RequestIDKey, rid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func newID() string {
	var b [12]byte
	_, _ = rand.Read(b[:])
	return base64.RawURLEncoding.EncodeToString(b[:])
}
