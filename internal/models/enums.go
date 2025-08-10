package models

// Status represents the lifecycle status of
// entities like tools, hubs, and virtual servers.
type Status string

const (
	StatusActive      Status = "ACTIVE"
	StatusDeactivated Status = "DEACTIVATED"
	StatusUnreachable Status = "UNREACHABLE"
)

// AuthType represents supported upstream auth mechanisms for hub connections.
type AuthType string

const (
	AuthTypeNone          AuthType = "none"
	AuthTypeBearer        AuthType = "bearer"
	AuthTypeCustomHeaders AuthType = "custom_headers"
)

// Role represents application user roles.
type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)
