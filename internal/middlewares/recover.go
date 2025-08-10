// Package middlewares contains HTTP middlewares used across the gateway.
package middlewares

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recover wraps a handler with panic recovery and logs the panic using slog.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("PANIC_LOG",
						"recover", rec,
						"stack", string(debug.Stack()))
					w.WriteHeader(http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
