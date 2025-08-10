// Package contextkey defines strongly-typed context keys used
// across the server.
package contextkey

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
