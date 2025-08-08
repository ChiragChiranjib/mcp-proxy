package models

import "time"

// MCPServer describes an upstream server catalog entry.
type MCPServer struct {
	ID          string    `gorm:"type:char(22);primaryKey"`
	Name        string    `gorm:"type:varchar(255);uniqueIndex"`
	URL         string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:varchar(255);default:''"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
