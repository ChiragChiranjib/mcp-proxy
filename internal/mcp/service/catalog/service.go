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

// NewService creates a new Cataloue Service.
func NewService(opts ...Option) *Service {
	s := &Service{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// List returns all catalog servers ordered by name.
func (s *Service) List(
	ctx context.Context,
) ([]m.MCPServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.ListCatalogServers(ctx)
}

// Add creates or updates a catalog server keyed by name.
func (s *Service) Add(ctx context.Context, srv m.MCPServer) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.CreateCatalogServer(ctx, srv)
}
