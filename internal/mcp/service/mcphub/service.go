// Package mcphub provides hub service implementation.
package mcphub

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	sqldb "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo/db"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/types"
)

// Service exposes hub operations.
type Service struct {
	q       *sqldb.Queries
	logger  *slog.Logger
	timeout time.Duration
}

// NewService creates a new hub Service.
func NewService(db *sql.DB, opts ...Option) *Service {
	s := &Service{q: sqldb.New(db)}
	for _, o := range opts {
		o.apply(s)
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
func (s *Service) Add(ctx context.Context, h types.HubServer) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.q.AddHubServer(ctx, sqldb.AddHubServerParams{
		ID: h.ID, UserID: h.UserID, McpServerID: h.MCServerID, Status: h.Status,
		Transport: h.Transport, Capabilities: h.Capabilities, AuthType: h.AuthType, AuthValue: h.AuthValue,
	})
}

// Get fetches a hub server by id.
func (s *Service) Get(ctx context.Context, id string) (types.HubServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	r, err := s.q.GetHubServer(ctx, id)
	if err != nil {
		return types.HubServer{}, err
	}
	return types.HubServer{ID: r.ID, UserID: r.UserID, MCServerID: r.McpServerID, Status: r.Status,
		Transport: r.Transport, Capabilities: r.Capabilities, AuthType: r.AuthType, AuthValue: r.AuthValue}, nil
}

// ListForUser returns hub servers for a user.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]types.HubServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	rows, err := s.q.ListUserHubServers(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]types.HubServer, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.HubServer{ID: r.ID, UserID: r.UserID, MCServerID: r.McpServerID, Status: r.Status,
			Transport: r.Transport, Capabilities: r.Capabilities, AuthType: r.AuthType, AuthValue: r.AuthValue})
	}
	return out, nil
}

// SetStatus updates hub server status.
func (s *Service) SetStatus(ctx context.Context, id string, status string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.q.UpdateHubServerStatus(ctx, sqldb.UpdateHubServerStatusParams{Status: status, ID: id})
}

// GetWithURL fetches hub with resolved server url and name.
func (s *Service) GetWithURL(ctx context.Context, id string) (types.HubServerWithURL, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	r, err := s.q.GetHubServerWithURL(ctx, id)
	if err != nil {
		return types.HubServerWithURL{}, err
	}
	return types.HubServerWithURL{
		HubServer: types.HubServer{ID: r.ID, UserID: r.UserID, MCServerID: r.McpServerID, Status: r.Status,
			Transport: r.Transport, Capabilities: r.Capabilities, AuthType: r.AuthType, AuthValue: r.AuthValue},
		ServerURL: r.ServerUrl, ServerName: r.ServerName,
	}, nil
}

// Delete removes a hub server.
func (s *Service) Delete(ctx context.Context, id string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.q.DeleteHubServer(ctx, id)
}
