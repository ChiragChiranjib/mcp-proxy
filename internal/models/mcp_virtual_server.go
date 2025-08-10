package models

import "time"

// MCPVirtualServer is a user-composed virtual server of tools.
type MCPVirtualServer struct {
	ID        string    `gorm:"type:char(22);primaryKey" json:"id"`
	UserID    string    `gorm:"type:char(22);index" json:"user_id"`
	Name      string    `gorm:"type:varchar(255);not null" json:"name"`
	Status    Status    `gorm:"type:varchar(30);not null" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName ...
func (MCPVirtualServer) TableName() string {
	return "mcp_virtual_servers"
}
