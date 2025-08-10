// Package middlewares contains HTTP middlewares used across the gateway.
package middlewares

import (
	"context"
	"net/http"

	ck "github.com/ChiragChiranjib/mcp-proxy/internal/contextkey"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
)

// Tag injects parameters in context for tracing.
func Tag() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get("X-Request-Id")
			if rid == "" {
				rid = idgen.NewID()
			}
			w.Header().Set("X-Request-Id", rid)
			ctx := context.WithValue(r.Context(), ck.RequestIDKey, rid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
