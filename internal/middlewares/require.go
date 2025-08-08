package middlewares

import (
	"net/http"
	"strings"

	ck "github.com/ChiragChiranjib/mcp-proxy/internal/contextkey"
)

// RequireAuthExcept returns a middleware that requires an authenticated user
// (as set in request context by other auth middlewares) for all requests,
// except when the request path starts with any of the provided prefixes.
func RequireAuthExcept(prefixes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := r.URL.Path
			for _, pref := range prefixes {
				if strings.HasPrefix(p, pref) {
					next.ServeHTTP(w, r)
					return
				}
			}
			if v := r.Context().Value(ck.UserIDKey); v == nil || v == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
