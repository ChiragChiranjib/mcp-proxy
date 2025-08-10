package models

import (
	"encoding/json"
	"time"
)

// MCPTool represents a discovered tool for a hub server.
type MCPTool struct {
	ID             string          `gorm:"type:char(22);primaryKey" json:"id"`
	UserID         string          `gorm:"type:char(22)" json:"user_id"`                                     //nolint:lll
	OriginalName   string          `gorm:"type:varchar(255);not null" json:"original_name"`                  //nolint:lll
	ModifiedName   string          `gorm:"type:varchar(255);" json:"modified_name"`                          //nolint:lll
	MCPHubServerID string          `gorm:"column:mcp_hub_server_id;type:char(22);" json:"mcp_hub_server_id"` //nolint:lll
	Description    string          `gorm:"type:text" json:"description"`
	InputSchema    json.RawMessage `gorm:"type:json" json:"input_schema"`
	Annotations    json.RawMessage `gorm:"type:json" json:"annotations"`
	Status         Status          `gorm:"type:varchar(30);not null" json:"status"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName ...
func (MCPTool) TableName() string { return "mcp_tools" }
