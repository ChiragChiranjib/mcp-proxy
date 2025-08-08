package middlewares

import (
	"context"
	"encoding/base64"
	"net/http"
	"strings"

	ck "github.com/ChiragChiranjib/mcp-proxy/internal/contextkey"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

// Auth parses the session cookie and injects user claims.
func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("session")
			if err == nil && c.Value != "" && jwtSecret != "" {
				token, _ := jwt.Parse(c.Value, func(t *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil })
				if token != nil && token.Valid {
					if claims, ok := token.Claims.(jwt.MapClaims); ok {
						uid, _ := claims["uid"].(string)
						email, _ := claims["email"].(string)
						role, _ := claims["role"].(string)
						if uid != "" {
							ctx := context.WithValue(r.Context(), ck.UserIDKey, uid)
							if email != "" {
								ctx = context.WithValue(ctx, ck.UserEmailKey, email)
							}
							if role != "" {
								ctx = context.WithValue(ctx, ck.UserRoleKey, role)
							}
							r = r.WithContext(ctx)
						}
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// BasicAuth inspects the Authorization header for Basic scheme and, if present, validates
// against the configured username/password. On success, it injects uid/email and role into context.
// It is non-blocking: requests without Authorization header continue as anonymous.
func BasicAuth(expectedUsername, expectedPassword, adminUserID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if !strings.HasPrefix(authz, "Basic ") {
				next.ServeHTTP(w, r)
				return
			}
			b64 := strings.TrimPrefix(authz, "Basic ")
			raw, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(string(raw), ":", 2)
			if len(parts) != 2 {
				w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			username := parts[0]
			password := parts[1]
			if username != expectedUsername || password != expectedPassword {
				w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			role := string(m.RoleUser)
			if adminUserID != "" && username == adminUserID {
				role = string(m.RoleAdmin)
			}
			ctx := context.WithValue(r.Context(), ck.UserIDKey, username)
			ctx = context.WithValue(ctx, ck.UserEmailKey, username)
			ctx = context.WithValue(ctx, ck.UserRoleKey, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
