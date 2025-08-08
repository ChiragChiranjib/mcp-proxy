// Package log provides a small wrapper for slog configuration.
package log

import (
	"log/slog"
	"os"
)

// Options controls logger configuration.
type Options struct {
	Level slog.Level
}

// New returns a JSON slog logger writing to stdout.
func New(options Options) *slog.Logger {
	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: options.Level})
	return slog.New(h)
}
