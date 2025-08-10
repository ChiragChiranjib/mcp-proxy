package middlewares

import (
	"context"
	"encoding/base64"
	"net/http"
	"regexp"
	"strings"

	"log/slog"

	"github.com/ChiragChiranjib/mcp-proxy/internal/config"
	ck "github.com/ChiragChiranjib/mcp-proxy/internal/contextkey"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/user"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

// Auth parses the session cookie and injects user claims.
func Auth(
	logger *slog.Logger,
	jwtSecret string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			c, err := r.Cookie("session")
			if err != nil || c == nil || c.Value == "" {
				// No session cookie; proceed without auth context
				next.ServeHTTP(w, r)
				return
			}

			token, perr := jwt.Parse(c.Value, func(_ *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if perr != nil || token == nil || !token.Valid {
				logger.Error("AUTH_JWT_PARSE_ERROR", "error", perr)
				next.ServeHTTP(w, r)
				return
			}

			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				uid, _ := claims["uid"].(string)
				email, _ := claims["email"].(string)
				role, _ := claims["role"].(string)

				ctx := context.WithValue(r.Context(), ck.UserIDKey, uid)
				ctx = context.WithValue(ctx, ck.UserEmailKey, email)
				ctx = context.WithValue(ctx, ck.UserRoleKey, role)
				r = r.WithContext(ctx)

				logger.Info("AUTH_JWT_OK", "uid", uid, "role", role)
			}

			next.ServeHTTP(w, r)
		})
	}
}

var skipRoutesForBasicAuth = []string{
	"/live",
	"/ready",
	"api/auth",
}

// mcpPathRE matches /servers/{22-char-id}/mcp exactly, where the id is
// URL-safe and can include hyphen or underscore in the last chars.
var mcpPathRE = regexp.MustCompile(
	`^/servers/[A-Za-z0-9_-]{22}/mcp/?$`,
)

// BasicAuth inspects the Authorization header for Basic scheme and,
// if present, validates against the configured username/password.
func BasicAuth(
	logger *slog.Logger,
	creds config.SecurityConfig,
	userService *user.Service,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p := r.URL.Path
			// Skip Basic auth for MCP streamable endpoint with 22-char id
			if mcpPathRE.MatchString(p) {
				next.ServeHTTP(w, r)
				return
			}
			for _, route := range skipRoutesForBasicAuth {
				if strings.HasPrefix(p, route) {
					next.ServeHTTP(w, r)
					return
				}
			}

			authz := r.Header.Get("Authorization")
			authToken := strings.SplitN(authz, " ", 2)
			if len(authToken) != 2 {
				logger.Error("BASIC_AUTH_INVALID_AUTH_HEADER")
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}

			raw, err := base64.StdEncoding.DecodeString(authToken[1])
			if err != nil {
				w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				logger.Error("BASIC_AUTH_DECODE_ERROR")
				return
			}
			parts := strings.SplitN(string(raw), ":", 2)
			if len(parts) != 2 {
				w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				logger.Error("BASIC_AUTH_FORMAT_ERROR")
				return
			}
			username := parts[0]
			password := parts[1]
			if username != creds.BasicUsername || password != creds.BasicPassword {
				w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				logger.Error("BASIC_AUTH_UNAUTHORIZED", "username", username)
				return
			}

			// Check if username exists in the DB
			userEntity, err := userService.FindUserByUserName(r.Context(), username)
			if err != nil || userEntity.Role != string(m.RoleAdmin) {
				http.Error(w, "unauthorized user", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ck.UserIDKey, userEntity.ID)
			ctx = context.WithValue(ctx, ck.UserEmailKey, userEntity.Username)
			ctx = context.WithValue(ctx, ck.UserRoleKey, userEntity.Role)

			logger.Info("BASIC_AUTH_SUCCESS",
				"username", userEntity.Username,
				"user_id", userEntity.ID,
				"role", userEntity.Role,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
