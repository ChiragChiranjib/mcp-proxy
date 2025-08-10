package catalog

import (
	"context"
	"log/slog"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
)

// Option configures the Service (functional options).
type Option func(*Service)

// WithLogger sets the logger for the service.
func WithLogger(l *slog.Logger) Option {
	return func(s *Service) { s.logger = l }
}

// WithRepo injects the GORM repo.
func WithRepo(r *repo.Repo) Option {
	return func(s *Service) {
		s.repo = r
	}
}

func (s *Service) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, s.timeout)
}
