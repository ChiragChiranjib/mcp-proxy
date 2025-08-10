package models

import "time"

// ToolVirtualServer is the pivot between tools and virtual servers.
type ToolVirtualServer struct {
	MCPVirtualServerID string    `gorm:"column:mcp_virtual_server_id;type:char(22);primaryKey"` //nolint:lll
	ToolID             string    `gorm:"type:char(22);primaryKey"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}

// TableName ...
func (ToolVirtualServer) TableName() string { return "tools_virtual_servers" }
