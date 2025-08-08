// Package virtualmcp provides the virtual server service implementation.
package virtualmcp

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	sqldb "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo/db"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/types"
)

// Service exposes virtual server operations.
type Service struct {
	q       *sqldb.Queries
	logger  *slog.Logger
	timeout time.Duration
}

// NewService creates a new virtual server Service.
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

// GetTools returns tools attached to a virtual server.
func (s *Service) GetTools(ctx context.Context, vsID string) ([]types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	rows, err := s.q.ListToolsForVirtualServer(ctx, vsID)
	if err != nil {
		return nil, err
	}
	out := make([]types.Tool, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.Tool{ID: r.ID, UserID: r.UserID, OriginalName: r.OriginalName, ModifiedName: r.ModifiedName,
			HubServerID: r.McpHubServerID, InputSchema: r.InputSchema, Annotations: r.Annotations, Status: r.Status})
	}
	return out, nil
}

// Create creates a new virtual server for a user.
func (s *Service) Create(ctx context.Context, userID string) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	id := "vs_" + idgen.NewID()
	if err := s.q.CreateVirtualServer(ctx, sqldb.CreateVirtualServerParams{ID: id, UserID: userID, Status: "ACTIVE"}); err != nil {
		return "", err
	}
	return id, nil
}

// ListForUser lists virtual servers for a user.
func (s *Service) ListForUser(ctx context.Context, userID string) ([]types.VirtualServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	rows, err := s.q.ListVirtualServersForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]types.VirtualServer, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.VirtualServer{ID: r.ID, UserID: r.UserID, Status: r.Status})
	}
	return out, nil
}

// SetStatus updates virtual server status.
func (s *Service) SetStatus(ctx context.Context, id string, status string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.q.UpdateVirtualServerStatus(ctx, sqldb.UpdateVirtualServerStatusParams{Status: status, ID: id})
}

// ReplaceTools replaces tool set for a virtual server (capped at 50).
func (s *Service) ReplaceTools(ctx context.Context, vsID string, toolIDs []string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	if err := s.q.ReplaceVirtualServerTools(ctx, vsID); err != nil {
		return err
	}
	if len(toolIDs) > 50 {
		toolIDs = toolIDs[:50]
	}
	for _, tid := range toolIDs {
		if err := s.q.AddVirtualServerTool(ctx, sqldb.AddVirtualServerToolParams{McpVirtualServerID: vsID, ToolID: tid}); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes a virtual server.
func (s *Service) Delete(ctx context.Context, id string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.q.DeleteVirtualServer(ctx, id)
}
