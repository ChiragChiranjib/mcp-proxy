// Package tool provides the Tool service for tool operations.
package tool

import (
	"log/slog"
	"time"
)

// Option configures the tool Service.
// Option configures the Tool service.
type Option interface{ apply(*Service) }

type loggerOption struct{ l *slog.Logger }

func (o loggerOption) apply(s *Service) { s.logger = o.l }

type timeoutOption struct{ d time.Duration }

func (o timeoutOption) apply(s *Service) { s.timeout = o.d }

// WithLogger sets a logger.
func WithLogger(l *slog.Logger) Option { return loggerOption{l: l} }

// WithTimeout sets a per-call timeout.
func WithTimeout(d time.Duration) Option { return timeoutOption{d: d} }
