package models

import "time"

// User represents an application user.
type User struct {
	ID        string    `gorm:"type:char(22);primaryKey"`
	Username  string    `gorm:"type:varchar(255);uniqueIndex"`
	Role      string    `gorm:"type:varchar(50);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName ...
func (User) TableName() string { return "users" }
