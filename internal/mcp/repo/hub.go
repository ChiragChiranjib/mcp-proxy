package repo

import (
	"context"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// CreateMCPHubServer ...
func (r *Repo) CreateMCPHubServer(
	ctx context.Context, h m.MCPHubServer) error {
	return r.WithContext(ctx).Create(&h).Error
}

// GetHubServerByID ...
func (r *Repo) GetHubServerByID(
	ctx context.Context, id string) (m.MCPHubServer, error) {
	var h m.MCPHubServer
	err := r.WithContext(ctx).
		Where("id = ?", id).
		Take(&h).Error
	return h, err
}

// HubWithURL ...
type HubWithURL struct {
	m.MCPHubServer
	ServerURL  string
	ServerName string
}

// GetHubServerWithURL ...
func (r *Repo) GetHubServerWithURL(
	ctx context.Context, id string) (m.MCPHubServerAggregate, error) {
	var out m.MCPHubServerAggregate
	err := r.WithContext(ctx).
		Table("mcp_hub_servers h").
		Select("h.id, h.user_id, h.mcp_server_id, h.status, "+
			"h.auth_type, h.auth_value, h.created_at, h.updated_at, "+
			"s.name AS name, s.url AS url, s.description AS description, "+
			"s.capabilities AS capabilities, s.transport AS transport, s.access_type AS access_type").
		Joins("JOIN mcp_servers s ON s.id = h.mcp_server_id").
		Where("h.id = ?", id).Scan(&out).Error
	return out, err
}

// ListUserHubMCPServers ...
func (r *Repo) ListUserHubMCPServers(
	ctx context.Context, userID string) ([]m.MCPHubServerAggregate, error) {
	var rows []m.MCPHubServerAggregate
	err := r.WithContext(ctx).
		Table("mcp_hub_servers h").
		Select("h.id, h.user_id, h.mcp_server_id, h.status, "+
			"h.auth_type, h.auth_value, h.created_at, h.updated_at, "+
			"s.name AS name, s.url AS url, s.description AS description, "+
			"s.capabilities AS capabilities, s.transport AS transport, s.access_type AS access_type").
		Joins("JOIN mcp_servers s ON s.id = h.mcp_server_id").
		Where("h.user_id = ?", userID).
		Scan(&rows).Error
	return rows, err
}

// UpdateHubServerStatus ...
func (r *Repo) UpdateHubServerStatus(
	ctx context.Context, id, status string) error {
	return r.WithContext(ctx).
		Table("mcp_hub_servers").
		Where("id = ?", id).
		Update("status", status).Error
}

// DeleteHubServer ...
func (r *Repo) DeleteHubServer(
	ctx context.Context, id string) error {
	return r.WithContext(ctx).
		Delete(&m.MCPHubServer{ID: id}).Error
}

// GetHubServerByServerAndUser gets a hub server by server ID and user ID.
func (r *Repo) GetHubServerByServerAndUser(
	ctx context.Context, serverID, userID string) (m.MCPHubServerAggregate, error) {
	var result m.MCPHubServerAggregate
	err := r.WithContext(ctx).
		Table("mcp_hub_servers h").
		Select(`
			h.id, h.user_id, h.mcp_server_id, h.status, h.auth_type, h.auth_value,
			h.created_at, h.updated_at,
			s.name AS name, s.url AS url, s.description AS description,
			s.capabilities AS capabilities, s.transport AS transport, s.access_type AS access_type
		`).
		Joins("JOIN mcp_servers s ON h.mcp_server_id = s.id").
		Where("h.mcp_server_id = ? AND h.user_id = ?", serverID, userID).
		Take(&result).Error
	return result, err
}
