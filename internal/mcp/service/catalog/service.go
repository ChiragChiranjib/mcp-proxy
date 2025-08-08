// Package catalog provides access to the MCP servers catalog.
package catalog

import (
	"context"
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// Service exposes catalog operations.
type Service struct {
	repo    *repo.Repo
	logger  *slog.Logger
	timeout time.Duration
}

// Option configures the Service (functional options).
type Option func(*Service)

// WithLogger sets the logger for the service.
func WithLogger(l *slog.Logger) Option { return func(s *Service) { s.logger = l } }

// WithRepo injects the GORM repo.
func WithRepo(r *repo.Repo) Option { return func(s *Service) { s.repo = r } }

// NewService creates a new hub Service.
func NewService(opts ...Option) *Service {
	s := &Service{}
	for _, o := range opts {
		o(s)
	}
	return s
}

func (s *Service) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, s.timeout)
}

// List returns all catalog servers ordered by name.
func (s *Service) List(
	ctx context.Context,
) ([]m.MCPServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	var rows []m.MCPServer
	if err := s.repo.WithContext(ctx).
		Order("name").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// Add creates or updates a catalog server keyed by name.
func (s *Service) Add(ctx context.Context, srv m.MCPServer) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.WithContext(ctx).Create(&srv).Error
}
