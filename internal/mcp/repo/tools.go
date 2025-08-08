package repo

import (
	"context"

	"github.com/ChiragChiranjib/mcp-proxy/internal/models"
	"gorm.io/gorm/clause"
)

// UpsertTool inserts or updates a tool.
func (r *Repo) UpsertTool(ctx context.Context, t models.MCPTool) error {
	return r.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "modified_name"}},
		DoUpdates: clause.Assignments(map[string]any{
			"input_schema": t.InputSchema,
			"annotations":  t.Annotations,
			"status":       t.Status,
		}),
	}).Create(&t).Error
}
