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

// Update modifies URL and/or description of a catalog server.
func (s *Service) Update(
	ctx context.Context,
	id string,
	url string,
	description string,
) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.UpdateCatalogServerURLDesc(ctx, id, url, description)
}

// UpdateCapabilities modifies capabilities and transport of a catalog server.
func (s *Service) UpdateCapabilities(
	ctx context.Context,
	id string,
	capabilities []byte,
	transport string,
) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.UpdateCatalogServerCapabilities(ctx, id, capabilities, transport)
}

// GetByID returns a catalog server by ID.
func (s *Service) GetByID(
	ctx context.Context,
	id string,
) (m.MCPServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.GetCatalogServerByID(ctx, id)
}

// ListPublic returns all public catalog servers.
func (s *Service) ListPublic(
	ctx context.Context,
) ([]m.MCPServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.ListPublicCatalogServers(ctx)
}

// ListPrivate returns all private catalog servers.
func (s *Service) ListPrivate(
	ctx context.Context,
) ([]m.MCPServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.ListPrivateCatalogServers(ctx)
}
