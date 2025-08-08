// Package tool provides the Tool service for tool operations.
package tool

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	sqldb "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo/db"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/types"
)

// Service provides tool operations backed by sqlc queries.
type Service struct {
	q       *sqldb.Queries
	logger  *slog.Logger
	timeout time.Duration
}

func (s *Service) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, s.timeout)
}

// NewService creates a tool Service (sqlc-backed, storage-agnostic API).
func NewService(db *sql.DB, opts ...Option) *Service {
	s := &Service{q: sqldb.New(db)}
	for _, o := range opts {
		o.apply(s)
	}
	return s
}

// ListForVirtualServer returns tools for a virtual server.
func (s *Service) ListForVirtualServer(ctx context.Context, vsID string) ([]types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	rows, err := s.q.ListToolsForVirtualServer(ctx, vsID)
	if err != nil {
		return nil, err
	}
	out := make([]types.Tool, 0, len(rows))
	for _, r := range rows {
		out = append(out, types.Tool{
			ID: r.ID, UserID: r.UserID,
			OriginalName: r.OriginalName, ModifiedName: r.ModifiedName,
			HubServerID: r.McpHubServerID,
			InputSchema: r.InputSchema, Annotations: r.Annotations,
			Status: r.Status,
		})
	}
	return out, nil
}

// SetStatus updates a tool status by id.
func (s *Service) SetStatus(ctx context.Context, id string, status string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.q.UpdateToolStatus(ctx, sqldb.UpdateToolStatusParams{Status: status, ID: id})
}

// Upsert inserts or updates a tool record.
func (s *Service) Upsert(ctx context.Context, t types.Tool) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	return s.q.UpsertTool(ctx, sqldb.UpsertToolParams{
		ID: t.ID, UserID: t.UserID,
		OriginalName: t.OriginalName, ModifiedName: t.ModifiedName,
		McpHubServerID: t.HubServerID, InputSchema: t.InputSchema,
		Annotations: t.Annotations, Status: t.Status,
	})
}

// GetByModifiedName returns a tool by user and modified name.
func (s *Service) GetByModifiedName(ctx context.Context, userID, modified string) (types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	r, err := s.q.GetToolByModifiedName(ctx, sqldb.GetToolByModifiedNameParams{UserID: userID, ModifiedName: modified})
	if err != nil {
		return types.Tool{}, err
	}
	return types.Tool{ID: r.ID, UserID: r.UserID, OriginalName: r.OriginalName, ModifiedName: r.ModifiedName,
		HubServerID: r.McpHubServerID, InputSchema: r.InputSchema, Annotations: r.Annotations, Status: r.Status}, nil
}

// ListForUserFiltered filters tools by hub, status, and query.
func (s *Service) ListForUserFiltered(ctx context.Context, userID, hubServerID, status, q string) ([]types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	rows, err := s.q.ListToolsForUserFiltered(ctx, sqldb.ListToolsForUserFilteredParams{
		UserID:  userID,
		Column2: hubServerID, McpHubServerID: hubServerID,
		Column4: status, Status: status,
		Column6: q, CONCAT: q, CONCAT_2: q,
	})
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

// ListActiveForHub returns active tools for a hub server.
func (s *Service) ListActiveForHub(ctx context.Context, hubServerID string) ([]types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	rows, err := s.q.ListActiveToolsForHub(ctx, hubServerID)
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
