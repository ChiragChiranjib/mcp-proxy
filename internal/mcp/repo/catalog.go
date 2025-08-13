package repo

import (
	"context"
	"errors"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// ListCatalogServers returns all catalog servers ordered by name.
func (r *Repo) ListCatalogServers(ctx context.Context) ([]m.MCPServer, error) {
	var rows []m.MCPServer
	err := r.WithContext(ctx).
		Order("name").
		Find(&rows).Error
	return rows, err
}

// CreateCatalogServer inserts a catalog server record.
func (r *Repo) CreateCatalogServer(ctx context.Context, srv m.MCPServer) error {
	return r.WithContext(ctx).Create(&srv).Error
}

// GetCatalogServerByID returns a catalog server by id.
func (r *Repo) GetCatalogServerByID(ctx context.Context, id string) (m.MCPServer, error) {
	var srv m.MCPServer
	err := r.WithContext(ctx).
		Where("id = ?", id).
		Take(&srv).Error
	return srv, err
}

// UpdateCatalogServerURLDesc updates URL and/or description for a catalog server.
// If both url and description are empty, it returns an error.
func (r *Repo) UpdateCatalogServerURLDesc(
	ctx context.Context,
	id string,
	url string,
	description string,
) error {
	updates := map[string]any{}
	if url != "" {
		updates["url"] = url
	}
	if description != "" {
		updates["description"] = description
	}
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}
	return r.WithContext(ctx).
		Model(&m.MCPServer{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateCatalogServerCapabilities updates capabilities and transport for a catalog server.
func (r *Repo) UpdateCatalogServerCapabilities(
	ctx context.Context,
	id string,
	capabilities []byte,
	transport string,
) error {
	updates := map[string]any{}
	if capabilities != nil {
		updates["capabilities"] = capabilities
	}
	if transport != "" {
		updates["transport"] = transport
	}
	if len(updates) == 0 {
		return errors.New("no fields to update")
	}
	return r.WithContext(ctx).
		Model(&m.MCPServer{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ListPublicCatalogServers returns all catalog servers with access_type = 'public'.
func (r *Repo) ListPublicCatalogServers(ctx context.Context) ([]m.MCPServer, error) {
	var rows []m.MCPServer
	err := r.WithContext(ctx).
		Where("access_type = ?", m.AccessTypePublic).
		Order("name").
		Find(&rows).Error
	return rows, err
}

// ListPrivateCatalogServers returns all catalog servers with access_type = 'private'.
func (r *Repo) ListPrivateCatalogServers(ctx context.Context) ([]m.MCPServer, error) {
	var rows []m.MCPServer
	err := r.WithContext(ctx).
		Where("access_type = ?", m.AccessTypePrivate).
		Order("name").
		Find(&rows).Error
	return rows, err
}
