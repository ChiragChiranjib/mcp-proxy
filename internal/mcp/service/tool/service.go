// Package tool provides the Tool service for tool operations.
package tool

import (
	"context"
	"log/slog"
	"time"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/types"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// Service provides tool operations backed by the GORM repo.
type Service struct {
	repo    *repo.Repo
	logger  *slog.Logger
	timeout time.Duration
}

func (s *Service) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, s.timeout)
}

// NewService creates a Tool service.
func NewService(opts ...Option) *Service {
	// placeholder, use WithRepo to inject
	s := &Service{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// ListForVirtualServer returns tools for a virtual server.
func (s *Service) ListForVirtualServer(ctx context.Context, vsID string) ([]types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	// Simplified: query via GORM join
	var tools []m.MCPTool
	err := s.repo.WithContext(ctx).
		Table("mcp_tools").
		Joins("JOIN tools_virtual_servers tvs ON tvs.tool_id = mcp_tools.id").
		Where("tvs.mcp_virtual_server_id = ?", vsID).
		Find(&tools).Error
	if err != nil {
		return nil, err
	}
	out := make([]types.Tool, 0, len(tools))
	for _, r := range tools {
		out = append(out, types.Tool{
			ID: r.ID, UserID: r.UserID,
			OriginalName: r.OriginalName, ModifiedName: r.ModifiedName,
			HubServerID: r.MCPHubServerID,
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
	return s.repo.WithContext(ctx).Table("mcp_tools").Where("id = ?", id).Update("status", status).Error
}

// Upsert inserts or updates a tool record.
func (s *Service) Upsert(ctx context.Context, t types.Tool) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	mt := m.MCPTool{ID: t.ID, UserID: t.UserID, OriginalName: t.OriginalName, ModifiedName: t.ModifiedName, MCPHubServerID: t.HubServerID, InputSchema: t.InputSchema, Annotations: t.Annotations, Status: t.Status}
	return s.repo.UpsertTool(ctx, mt)
}

// GetByModifiedName returns a tool by user and modified name.
func (s *Service) GetByModifiedName(ctx context.Context, userID, modified string) (types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	var r m.MCPTool
	err := s.repo.WithContext(ctx).Where("user_id = ? AND modified_name = ?", userID, modified).Take(&r).Error
	if err != nil {
		return types.Tool{}, err
	}
	return types.Tool{ID: r.ID, UserID: r.UserID, OriginalName: r.OriginalName, ModifiedName: r.ModifiedName,
		HubServerID: r.MCPHubServerID, InputSchema: r.InputSchema, Annotations: r.Annotations, Status: r.Status}, nil
}

// ListForUserFiltered filters tools by hub, status, and query.
func (s *Service) ListForUserFiltered(ctx context.Context, userID, hubServerID, status, q string) ([]types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	qdb := s.repo.WithContext(ctx).Table("mcp_tools").Where("user_id = ?", userID)
	if hubServerID != "" {
		qdb = qdb.Where("mcp_hub_server_id = ?", hubServerID)
	}
	if status != "" {
		qdb = qdb.Where("status = ?", status)
	}
	if q != "" {
		qdb = qdb.Where("modified_name LIKE ? OR original_name LIKE ?", "%"+q+"%", "%"+q+"%")
	}
	var tools []m.MCPTool
	if err := qdb.Order("modified_name").Find(&tools).Error; err != nil {
		return nil, err
	}
	out := make([]types.Tool, 0, len(tools))
	for _, r := range tools {
		out = append(out, types.Tool{ID: r.ID, UserID: r.UserID, OriginalName: r.OriginalName, ModifiedName: r.ModifiedName,
			HubServerID: r.MCPHubServerID, InputSchema: r.InputSchema, Annotations: r.Annotations, Status: r.Status})
	}
	return out, nil
}

// ListActiveForHub returns active tools for a hub server.
func (s *Service) ListActiveForHub(ctx context.Context, hubServerID string) ([]types.Tool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	var tools []m.MCPTool
	if err := s.repo.WithContext(ctx).Where("mcp_hub_server_id = ? AND status = 'ACTIVE'", hubServerID).Find(&tools).Error; err != nil {
		return nil, err
	}
	out := make([]types.Tool, 0, len(tools))
	for _, r := range tools {
		out = append(out, types.Tool{ID: r.ID, UserID: r.UserID, OriginalName: r.OriginalName, ModifiedName: r.ModifiedName,
			HubServerID: r.MCPHubServerID, InputSchema: r.InputSchema, Annotations: r.Annotations, Status: r.Status})
	}
	return out, nil
}
