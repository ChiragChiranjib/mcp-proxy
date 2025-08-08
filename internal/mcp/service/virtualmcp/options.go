// Package virtualmcp provides the virtual server service implementation.
package virtualmcp

import (
	"log/slog"
	"time"
)

// Option configures the virtual Service.
type Option interface{ apply(*Service) }

type loggerOption struct{ l *slog.Logger }

func (o loggerOption) apply(s *Service) { s.logger = o.l }

type timeoutOption struct{ d time.Duration }

func (o timeoutOption) apply(s *Service) { s.timeout = o.d }

// WithLogger sets a logger.
func WithLogger(l *slog.Logger) Option { return loggerOption{l: l} }

// WithTimeout sets a per-call timeout.
func WithTimeout(d time.Duration) Option { return timeoutOption{d: d} }
