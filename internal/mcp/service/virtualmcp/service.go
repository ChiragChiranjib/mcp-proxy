// Package virtualmcp provides the virtual server service implementation.
package virtualmcp

import (
	"context"
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// Service exposes virtual server operations.
type Service struct {
	repo    *repo.Repo
	logger  *slog.Logger
	timeout time.Duration
}

// NewService creates a new virtual server Service.
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

// GetTools returns tools attached to a virtual server.
func (s *Service) GetTools(
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

// Create creates a new virtual server for a user.
func (s *Service) Create(ctx context.Context, userID string) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	id := "vs_" + idgen.NewID()
	if err := s.repo.CreateVirtualServer(ctx, m.MCPVirtualServer{
		ID:     id,
		UserID: userID,
		Status: m.StatusActive,
	}); err != nil {
		return "", err
	}
	return id, nil
}

// ListForUser lists virtual servers for a user.
func (s *Service) ListForUser(
	ctx context.Context,
	userID string,
) ([]m.MCPVirtualServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	rows, err := s.repo.ListVirtualServersForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// SetStatus updates virtual server status.
func (s *Service) SetStatus(ctx context.Context, id string, status string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.UpdateVirtualServerStatus(ctx, id, status)
}

// ReplaceTools replaces tool set for a virtual server (capped at 50).
func (s *Service) ReplaceTools(
	ctx context.Context,
	vsID string,
	toolIDs []string,
) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	if err := s.repo.ReplaceVirtualServerTools(ctx, vsID); err != nil {
		return err
	}
	if len(toolIDs) > 50 {
		toolIDs = toolIDs[:50]
	}
	for _, tid := range toolIDs {
		if err := s.repo.AddVirtualServerTool(ctx, vsID, tid); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes a virtual server.
func (s *Service) Delete(ctx context.Context, id string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.repo.DeleteVirtualServer(ctx, id)
}
