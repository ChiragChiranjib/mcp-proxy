// Package virtualmcp provides the virtual server service implementation.
package virtualmcp

import (
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
)

// Option configures the virtual Service (functional options).
type Option func(*Service)

// WithLogger sets a logger.
func WithLogger(l *slog.Logger) Option { return func(s *Service) { s.logger = l } }

// WithTimeout sets a per-call timeout.
func WithTimeout(d time.Duration) Option { return func(s *Service) { s.timeout = d } }

// WithRepo injects the GORM repo for virtual service
func WithRepo(r *repo.Repo) Option { return func(s *Service) { s.repo = r } }
