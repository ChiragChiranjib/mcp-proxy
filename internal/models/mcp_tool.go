package models

import "time"

// MCPTool represents a discovered tool for a hub server.
type MCPTool struct {
	ID             string    `gorm:"type:char(22);primaryKey" json:"id"`
	UserID         string    `gorm:"type:char(22);index:user_mod,priority:1" json:"user_id"`
	OriginalName   string    `gorm:"type:varchar(255);not null" json:"original_name"`
	ModifiedName   string    `gorm:"type:varchar(255);index:user_mod,priority:2" json:"modified_name"`
	MCPHubServerID string    `gorm:"column:mcp_hub_server_id;type:char(22);index" json:"mcp_hub_server_id"`
	Description    string    `gorm:"type:text" json:"description"`
	InputSchema    []byte    `gorm:"type:json" json:"input_schema"`
	Annotations    []byte    `gorm:"type:json" json:"annotations"`
	Status         Status    `gorm:"type:varchar(30);not null" json:"status"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MCPTool) TableName() string { return "mcp_tools" }
