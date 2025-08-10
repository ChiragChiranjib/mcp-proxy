package repo

import (
	"context"

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
