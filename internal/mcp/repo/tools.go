package repo

import (
	"context"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
	"gorm.io/gorm/clause"
)

// UpsertTool inserts or updates a tool using the new unique constraint.
func (r *Repo) UpsertTool(ctx context.Context, t m.MCPTool) error {
	return r.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "mcp_server_id"}, {Name: "user_id"}, {Name: "original_name"}},
		DoUpdates: clause.Assignments(map[string]any{
			"modified_name": t.ModifiedName,
			"description":   t.Description,
			"input_schema":  t.InputSchema,
			"annotations":   t.Annotations,
			"status":        t.Status,
		}),
	}).Create(&t).Error
}

// ListToolsForVirtualServer returns tools joined via tools_virtual_servers
// for a vs id.
func (r *Repo) ListToolsForVirtualServer(
	ctx context.Context, vsID string) ([]m.MCPTool, error) {
	var tools []m.MCPTool
	err := r.WithContext(ctx).
		Table("mcp_tools").
		Joins("JOIN tools_virtual_servers tvs ON tvs.tool_id = mcp_tools.id").
		Where("tvs.mcp_virtual_server_id = ?", vsID).
		Find(&tools).Error
	return tools, err
}

// ListUserToolsFiltered returns tools for a user filtered by server,
// status, and query. Includes both global tools (user_id=NULL) and user-specific tools.
func (r *Repo) ListUserToolsFiltered(
	ctx context.Context,
	userID,
	serverID,
	status,
	q string) ([]m.MCPTool, error) {
	return r.ListUserToolsFilteredWithHub(ctx, userID, serverID, "", status, q)
}

// ListUserToolsFilteredWithHub returns tools for a user filtered by server, hub, status, and query.
func (r *Repo) ListUserToolsFilteredWithHub(
	ctx context.Context,
	userID,
	serverID,
	hubServerID,
	status,
	q string) ([]m.MCPTool, error) {
	qdb := r.WithContext(ctx).Table("mcp_tools").
		Where("user_id IS NULL OR user_id = ?", userID)

	if serverID != "" {
		qdb = qdb.Where("mcp_server_id = ?", serverID)
	}
	if hubServerID != "" {
		qdb = qdb.Where("mcp_hub_server_id = ?", hubServerID)
	}
	if status != "" {
		qdb = qdb.Where("status = ?", status)
	}
	if q != "" {
		qdb = qdb.Where("modified_name LIKE ? OR original_name LIKE ?", "%"+q+"%", "%"+q+"%") //nolint:lll
	}
	var tools []m.MCPTool
	if err := qdb.Order("modified_name").Find(&tools).Error; err != nil {
		return nil, err
	}
	return tools, nil
}

// ListToolsForServer returns all tools for a specific server (both global and user-specific).
func (r *Repo) ListToolsForServer(
	ctx context.Context,
	serverID string,
	userID *string) ([]m.MCPTool, error) {
	qdb := r.WithContext(ctx).Where("mcp_server_id = ?", serverID)

	if userID == nil {
		// Only global tools
		qdb = qdb.Where("user_id IS NULL")
	} else {
		// Both global tools and user-specific tools
		qdb = qdb.Where("user_id IS NULL OR user_id = ?", *userID)
	}

	var tools []m.MCPTool
	err := qdb.Order("original_name").Find(&tools).Error
	return tools, err
}

// ListGlobalToolsForServer returns only global tools for a server.
func (r *Repo) ListGlobalToolsForServer(
	ctx context.Context,
	serverID string) ([]m.MCPTool, error) {
	var tools []m.MCPTool
	err := r.WithContext(ctx).
		Where("mcp_server_id = ? AND user_id IS NULL", serverID).
		Order("original_name").
		Find(&tools).Error
	return tools, err
}

// ListUserSpecificToolsForServer returns only user-specific tools for a server.
func (r *Repo) ListUserSpecificToolsForServer(
	ctx context.Context,
	serverID string,
	userID string) ([]m.MCPTool, error) {
	var tools []m.MCPTool
	err := r.WithContext(ctx).
		Where("mcp_server_id = ? AND user_id = ?", serverID, userID).
		Order("original_name").
		Find(&tools).Error
	return tools, err
}

// DeleteToolsForServer deletes all tools for a server (used when refreshing global tools).
func (r *Repo) DeleteToolsForServer(
	ctx context.Context,
	serverID string,
	userID *string) error {
	qdb := r.WithContext(ctx).Where("mcp_server_id = ?", serverID)

	if userID == nil {
		// Delete only global tools
		qdb = qdb.Where("user_id IS NULL")
	} else {
		// Delete only user-specific tools
		qdb = qdb.Where("user_id = ?", *userID)
	}

	return qdb.Delete(&m.MCPTool{}).Error
}

// GetActiveToolByID returns a tool by id only if it is ACTIVE.
func (r *Repo) GetActiveToolByID(
	ctx context.Context, id string) (m.MCPTool, error) {
	var t m.MCPTool
	err := r.WithContext(ctx).
		Where("id = ? AND status = 'ACTIVE'", id).
		Take(&t).Error
	return t, err
}

// CreateTools creates multiple tools in bulk.
func (r *Repo) CreateTools(ctx context.Context, tools []m.MCPTool) error {
	if len(tools) == 0 {
		return nil
	}
	return r.WithContext(ctx).Create(&tools).Error
}

// DeleteToolsByIDs deletes tools by their IDs.
func (r *Repo) DeleteToolsByIDs(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	return r.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&m.MCPTool{}).Error
}
