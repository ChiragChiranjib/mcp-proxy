package repo

import (
	"context"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

func (r *Repo) AddHubServer(ctx context.Context, h m.MCPHubServer) error {
	return r.WithContext(ctx).Create(&h).Error
}

func (r *Repo) GetHubServer(ctx context.Context, id string) (m.MCPHubServer, error) {
	var h m.MCPHubServer
	err := r.WithContext(ctx).Where("id = ?", id).Take(&h).Error
	return h, err
}

type HubWithURL struct {
	m.MCPHubServer
	ServerURL  string
	ServerName string
}

func (r *Repo) GetHubServerWithURL(ctx context.Context, id string) (HubWithURL, error) {
	var out HubWithURL
	err := r.WithContext(ctx).Table("mcp_hub_servers h").
		Select("h.*, s.url as server_url, s.name as server_name").
		Joins("JOIN mcp_servers s ON s.id = h.mcp_server_id").
		Where("h.id = ?", id).Scan(&out).Error
	return out, err
}

func (r *Repo) ListUserHubServers(ctx context.Context, userID string) ([]m.MCPHubServer, error) {
	var rows []m.MCPHubServer
	err := r.WithContext(ctx).Where("user_id = ?", userID).Find(&rows).Error
	return rows, err
}

func (r *Repo) UpdateHubServerStatus(ctx context.Context, id, status string) error {
	return r.WithContext(ctx).Table("mcp_hub_servers").Where("id = ?", id).Update("status", status).Error
}

func (r *Repo) DeleteHubServer(ctx context.Context, id string) error {
	return r.WithContext(ctx).Delete(&m.MCPHubServer{ID: id}).Error
}
