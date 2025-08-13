// Package tool provides the Tool service for tool operations.
package tool

import (
	"context"
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// Service provides tool operations backed by the GORM repo.
type Service struct {
	repo    *repo.Repo
	logger  *slog.Logger
	timeout time.Duration
}

func (s *Service) withTimeout(
	ctx context.Context) (context.Context, context.CancelFunc) {
	if s.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, s.timeout)
}

// NewService creates a Tool service.
func NewService(opts ...Option) *Service {
	s := &Service{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// ListForVirtualServer returns tools for a virtual server.
func (s *Service) ListForVirtualServer(
	ctx context.Context,
	vsID string,
) ([]m.MCPTool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	tools, err := s.repo.ListToolsForVirtualServer(ctx, vsID)
	if err != nil {
		return nil, err
	}
	return tools, nil
}

// SetStatus updates a tool status by id.
func (s *Service) SetStatus(
	ctx context.Context, id string, status string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.WithContext(ctx).
		Table("mcp_tools").
		Where("id = ?", id).
		Update("status", status).
		Error
}

// Upsert inserts or updates a tool record.
func (s *Service) Upsert(ctx context.Context, t m.MCPTool) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	if t.ID == "" {
		t.ID = idgen.NewID()
	}
	return s.repo.UpsertTool(ctx, t)
}

// GetByModifiedName returns a tool by user and modified name.
func (s *Service) GetByModifiedName(
	ctx context.Context,
	userID, modified string,
) (m.MCPTool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	var r m.MCPTool
	err := s.repo.WithContext(ctx).
		Where("user_id = ? AND modified_name = ?", userID, modified).
		Take(&r).Error
	if err != nil {
		return m.MCPTool{}, err
	}
	return r, nil
}

// ListForUserFiltered filters tools by server, hub, status, and query.
func (s *Service) ListForUserFiltered(
	ctx context.Context,
	userID, serverID, hubServerID, status, q string,
) ([]m.MCPTool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	tools, err := s.repo.ListUserToolsFilteredWithHub(ctx, userID, serverID, hubServerID, status, q)
	if err != nil {
		return nil, err
	}
	return tools, nil
}

// ListGlobalToolsForServer returns global tools for a specific server.
func (s *Service) ListGlobalToolsForServer(
	ctx context.Context,
	serverID string,
) ([]m.MCPTool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.ListGlobalToolsForServer(ctx, serverID)
}
