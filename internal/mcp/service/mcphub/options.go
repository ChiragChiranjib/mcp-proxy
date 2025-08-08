// Package mcphub provides hub service implementation.
package mcphub

import (
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
)

// Option configures the hub Service (functional options).
type Option func(*Service)

// WithLogger sets a logger.
func WithLogger(l *slog.Logger) Option { return func(s *Service) { s.logger = l } }

// WithTimeout sets a per-call timeout.
func WithTimeout(d time.Duration) Option { return func(s *Service) { s.timeout = d } }

// WithRepo injects the GORM repo.
func WithRepo(r *repo.Repo) Option { return func(s *Service) { s.repo = r } }
