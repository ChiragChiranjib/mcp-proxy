package models

import "time"

// MCPServer describes an upstream server catalog entry.
type MCPServer struct {
	ID          string    `gorm:"type:char(22);primaryKey" json:"id"`
	Name        string    `gorm:"type:varchar(255);uniqueIndex" json:"name"`
	URL         string    `gorm:"type:varchar(255);not null" json:"url"`
	Description string    `gorm:"type:varchar(255);default:''" json:"description"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MCPServer) TableName() string { return "mcp_servers" }
