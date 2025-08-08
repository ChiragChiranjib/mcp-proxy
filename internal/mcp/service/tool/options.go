// Package tool provides the Tool service for tool operations.
package tool

import (
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
)

// Option configures the Tool service (functional options).
type Option func(*Service)

// WithLogger sets a logger.
func WithLogger(l *slog.Logger) Option { return func(s *Service) { s.logger = l } }

// WithTimeout sets a per-call timeout.
func WithTimeout(d time.Duration) Option { return func(s *Service) { s.timeout = d } }

// WithRepo injects the GORM repo
func WithRepo(r *repo.Repo) Option { return func(s *Service) { s.repo = r } }
