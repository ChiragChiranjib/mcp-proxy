// Package mcphub provides hub service implementation.
package mcphub

import (
	"context"
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// Service exposes hub operations.
type Service struct {
	repo    *repo.Repo
	logger  *slog.Logger
	timeout time.Duration
}

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

// Add registers a new hub server.
func (s *Service) Add(ctx context.Context, h m.MCPHubServer) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	s.logger.Info("MCP_HUB_ADD_INIT", "id", h.ID, "user_id", h.UserID, "mcp_server_id", h.MCPServerID)
	err := s.repo.AddHubServer(ctx, h)
	if err != nil {
		s.logger.Error("MCP_HUB_ADD_ERROR", "error", err)
		return err
	}
	s.logger.Info("MCP_HUB_ADD_OK", "id", h.ID)
	return nil
}

// Get fetches a hub server by id.
func (s *Service) Get(ctx context.Context, id string) (m.MCPHubServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	s.logger.Info("MCP_HUB_GET_INIT", "id", id)
	r, err := s.repo.GetHubServer(ctx, id)
	if err != nil {
		s.logger.Error("MCP_HUB_GET_ERROR", "error", err)
		return m.MCPHubServer{}, err
	}
	s.logger.Info("MCP_HUB_GET_OK", "id", r.ID)
	return r, nil
}

// ListForUser returns hub servers for a user.
func (s *Service) ListForUser(
	ctx context.Context,
	userID string,
) ([]m.MCPHubServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	s.logger.Info("MCP_HUB_LIST_FOR_USER_INIT", "user_id", userID)
	rows, err := s.repo.ListUserHubServers(ctx, userID)
	if err != nil {
		s.logger.Error("MCP_HUB_LIST_FOR_USER_ERROR", "error", err)
		return nil, err
	}
	s.logger.Info("MCP_HUB_LIST_FOR_USER_OK", "count", len(rows))
	return rows, nil
}

// SetStatus updates hub server status.
func (s *Service) SetStatus(ctx context.Context, id string, status string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	s.logger.Info("MCP_HUB_SET_STATUS_INIT", "id", id, "status", status)
	err := s.repo.UpdateHubServerStatus(ctx, id, status)
	if err != nil {
		s.logger.Error("MCP_HUB_SET_STATUS_ERROR", "error", err)
		return err
	}
	s.logger.Info("MCP_HUB_SET_STATUS_OK", "id", id)
	return nil
}

// GetWithURL fetches hub with resolved server url and name.
func (s *Service) GetWithURL(ctx context.Context, id string) (repo.HubWithURL, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	s.logger.Info("MCP_HUB_GET_WITH_URL_INIT", "id", id)
	r, err := s.repo.GetHubServerWithURL(ctx, id)
	if err != nil {
		s.logger.Error("MCP_HUB_GET_WITH_URL_ERROR", "error", err)
		return repo.HubWithURL{}, err
	}
	s.logger.Info("MCP_HUB_GET_WITH_URL_OK", "id", r.ID)
	return r, nil
}

// Delete removes a hub server.
func (s *Service) Delete(ctx context.Context, id string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	s.logger.Info("MCP_HUB_DELETE_INIT", "id", id)
	err := s.repo.DeleteHubServer(ctx, id)
	if err != nil {
		s.logger.Error("MCP_HUB_DELETE_ERROR", "error", err)
		return err
	}
	s.logger.Info("MCP_HUB_DELETE_OK", "id", id)
	return nil
}
