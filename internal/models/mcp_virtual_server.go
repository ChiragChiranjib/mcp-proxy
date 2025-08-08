package models

import "time"

// MCPVirtualServer is a user-composed virtual server of tools.
type MCPVirtualServer struct {
	ID        string    `gorm:"type:char(22);primaryKey"`
	UserID    string    `gorm:"type:char(22);index"`
	Status    string    `gorm:"type:varchar(30);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (MCPVirtualServer) TableName() string { return "mcp_virtual_servers" }
