package models

import "time"

// MCPHubServer is a user-added upstream in the hub.
type MCPHubServer struct {
	ID           string    `gorm:"type:char(22);primaryKey" json:"id"`
	UserID       string    `gorm:"type:char(22);index" json:"user_id"`
	MCPServerID  string    `gorm:"column:mcp_server_id;type:char(22);index" json:"mcp_server_id"`
	Status       Status    `gorm:"type:varchar(30);not null" json:"status"`
	Transport    string    `gorm:"type:varchar(30);not null" json:"transport"`
	Capabilities []byte    `gorm:"type:json" json:"capabilities"`
	AuthType     AuthType  `gorm:"type:varchar(30);not null" json:"auth_type"`
	AuthValue    []byte    `gorm:"type:json" json:"auth_value"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (MCPHubServer) TableName() string { return "mcp_hub_servers" }
