// Package types defines shared domain structs used by services.
package types

import "encoding/json"

// Status represents the status of an entity.
type Status string

const (
	// StatusActive indicates an active entity.
	StatusActive Status = "ACTIVE"
	// StatusDeactivated indicates a disabled entity.
	StatusDeactivated Status = "DEACTIVATED"
	// StatusUnreachable indicates connectivity failures.
	StatusUnreachable Status = "UNREACHABLE"
)

// AuthType represents supported upstream auth mechanisms.
type AuthType string

const (
	// AuthNone disables upstream auth.
	AuthNone AuthType = "none"
	// AuthBearer uses bearer token.
	AuthBearer AuthType = "bearer"
	// AuthCustomHeader forwards custom headers.
	AuthCustomHeader AuthType = "custom_headers"
)

// HubServer models a registered upstream MCP hub server.
type HubServer struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	MCServerID   string          `json:"mcp_server_id"`
	Status       string          `json:"status"`
	Transport    string          `json:"transport"`
	Capabilities json.RawMessage `json:"capabilities"`
	AuthType     string          `json:"auth_type"`
	AuthValue    json.RawMessage `json:"auth_value"`
}

// HubServerWithURL provides resolved URL and name for a hub server.
type HubServerWithURL struct {
	HubServer
	ServerURL  string `json:"server_url"`
	ServerName string `json:"server_name"`
}

// Tool describes a proxied MCP tool registered in the system.
type Tool struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	OriginalName string          `json:"original_name"`
	ModifiedName string          `json:"modified_name"`
	HubServerID  string          `json:"hub_server_id"`
	InputSchema  json.RawMessage `json:"input_schema"`
	Annotations  json.RawMessage `json:"annotations"`
	Status       string          `json:"status"`
}

// VirtualServer represents a user-defined virtual server composed of tools.
type VirtualServer struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Status string `json:"status"`
}
