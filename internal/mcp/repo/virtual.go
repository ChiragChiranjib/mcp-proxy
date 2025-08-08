package repo

import (
	"context"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

func (r *Repo) CreateVirtualServer(ctx context.Context, vs m.MCPVirtualServer) error {
	return r.WithContext(ctx).Create(&vs).Error
}

func (r *Repo) ListVirtualServersForUser(ctx context.Context, userID string) ([]m.MCPVirtualServer, error) {
	var rows []m.MCPVirtualServer
	err := r.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error
	return rows, err
}

func (r *Repo) UpdateVirtualServerStatus(ctx context.Context, id, status string) error {
	return r.WithContext(ctx).Table("mcp_virtual_servers").Where("id = ?", id).Update("status", status).Error
}

func (r *Repo) ReplaceVirtualServerTools(ctx context.Context, vsID string) error {
	return r.WithContext(ctx).Where("mcp_virtual_server_id = ?", vsID).Delete(&m.ToolVirtualServer{}).Error
}

func (r *Repo) AddVirtualServerTool(ctx context.Context, vsID, toolID string) error {
	rec := m.ToolVirtualServer{MCPVirtualServerID: vsID, ToolID: toolID}
	return r.WithContext(ctx).Create(&rec).Error
}

func (r *Repo) DeleteVirtualServer(ctx context.Context, id string) error {
	return r.WithContext(ctx).Delete(&m.MCPVirtualServer{ID: id}).Error
}
