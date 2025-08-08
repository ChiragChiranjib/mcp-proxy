package repo

import (
	"context"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
	"gorm.io/gorm/clause"
)

// UpsertTool inserts or updates a tool.
func (r *Repo) UpsertTool(ctx context.Context, t m.MCPTool) error {
	return r.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "modified_name"}},
		DoUpdates: clause.Assignments(map[string]any{
			"input_schema": t.InputSchema,
			"annotations":  t.Annotations,
			"status":       t.Status,
		}),
	}).Create(&t).Error
}

// ListToolsForVirtualServer returns tools joined via tools_virtual_servers for a vs id.
func (r *Repo) ListToolsForVirtualServer(ctx context.Context, vsID string) ([]m.MCPTool, error) {
	var tools []m.MCPTool
	err := r.WithContext(ctx).
		Table("mcp_tools").
		Joins("JOIN tools_virtual_servers tvs ON tvs.tool_id = mcp_tools.id").
		Where("tvs.mcp_virtual_server_id = ?", vsID).
		Find(&tools).Error
	return tools, err
}

// ListUserToolsFiltered returns tools for a user filtered by hub, status, and query.
func (r *Repo) ListUserToolsFiltered(ctx context.Context, userID, hubServerID, status, q string) ([]m.MCPTool, error) {
	qdb := r.WithContext(ctx).Table("mcp_tools").Where("user_id = ?", userID)
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
	return tools, nil
}

// ListActiveToolsForHub returns active tools for a hub server.
func (r *Repo) ListActiveToolsForHub(ctx context.Context, hubServerID string) ([]m.MCPTool, error) {
	var tools []m.MCPTool
	if err := r.WithContext(ctx).Where("mcp_hub_server_id = ? AND status = 'ACTIVE'", hubServerID).Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}
