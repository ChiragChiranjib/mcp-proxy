package models

import "time"

// MCPTool represents a discovered tool for a hub server.
type MCPTool struct {
	ID             string    `gorm:"type:char(22);primaryKey"`
	UserID         string    `gorm:"type:char(22);index:user_mod,priority:1"`
	OriginalName   string    `gorm:"type:varchar(255);not null"`
	ModifiedName   string    `gorm:"type:varchar(255);index:user_mod,priority:2"`
	MCPHubServerID string    `gorm:"column:mcp_hub_server_id;type:char(22);index"`
	InputSchema    []byte    `gorm:"type:json"`
	Annotations    []byte    `gorm:"type:json"`
	Status         string    `gorm:"type:varchar(30);not null"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}

func (MCPTool) TableName() string { return "mcp_tools" }
