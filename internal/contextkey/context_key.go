// Package contextkey defines strongly-typed context keys used
// across the server.
package contextkey

import (
	"context"
)

// ContextKey is a string-based key type for request context values.
// Using a distinct type prevents collisions with other packages' keys.
type ContextKey string

const (
	// UserIDKey holds the authenticated user's ID in request context.
	UserIDKey ContextKey = "uid"

	// UserEmailKey holds the authenticated user's email in request context.
	UserEmailKey ContextKey = "email"

	// UserRoleKey holds the user's role, e.g., "admin" or "user".
	UserRoleKey ContextKey = "role"

	// RequestIDKey is the tracing request id key used by the
	// request id middleware.
	RequestIDKey ContextKey = "request_id"
)

// fromContext returns the string value for the given key if present.
func fromContext(ctx context.Context, key ContextKey) string {
	if v := ctx.Value(key); v != nil {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return ""
}

// GetUserIDFromContext returns the user id from context.
func GetUserIDFromContext(ctx context.Context) string {
	return fromContext(ctx, UserIDKey)
}

// GetUserEmailFromContext returns the user email from context.
func GetUserEmailFromContext(ctx context.Context) string {
	return fromContext(ctx, UserEmailKey)
}

// GetUserRoleFromContext returns the user role from context.
func GetUserRoleFromContext(ctx context.Context) string {
	return fromContext(ctx, UserRoleKey)
}

// GetRequestIDFromContext returns the request id from context.
func GetRequestIDFromContext(ctx context.Context) string {
	return fromContext(ctx, RequestIDKey)
}
