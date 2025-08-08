package models

import "time"

// MCPHubServer is a user-added upstream in the hub.
type MCPHubServer struct {
	ID           string    `gorm:"type:char(22);primaryKey"`
	UserID       string    `gorm:"type:char(22);index"`
	MCPServerID  string    `gorm:"column:mcp_server_id;type:char(22);index"`
	Status       string    `gorm:"type:varchar(30);not null"`
	Transport    string    `gorm:"type:varchar(30);not null"`
	Capabilities []byte    `gorm:"type:json"`
	AuthType     string    `gorm:"type:varchar(30);not null"`
	AuthValue    []byte    `gorm:"type:json"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (MCPHubServer) TableName() string { return "mcp_hub_servers" }
