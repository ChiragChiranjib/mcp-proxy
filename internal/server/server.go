// Package server wires dependencies and builds the HTTP handler container.
package server

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/catalog"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/tool"
	usersvc "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/user"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/virtualmcp"
	"github.com/ChiragChiranjib/mcp-proxy/internal/middlewares"
)

// Server holds the final HTTP handler for the app.
type Server struct {
	Handler http.Handler
}

// Deps enumerates dependencies required to assemble the server.
type Deps struct {
	Logger         *slog.Logger
	Tools          *tool.Service
	Hubs           *mcphub.Service
	Virtual        *virtualmcp.Service
	Catalog        *catalog.Service
	Encrypter      *encryptor.AESEncrypter
	GoogleClientID string
	JWTSecret      string
	UserService    *usersvc.Service
	// Optional Basic auth
	BasicUsername string
	BasicPassword string
	// Admin user id to assign admin role
	AdminUserID string
}

// Config holds HTTP wiring configuration.
type Config struct {
	Router      *mux.Router
	Middlewares []func(http.Handler) http.Handler
	MCPMount    string
	AdminPrefix string
	Health      bool
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MCPMount:    "/servers/{virtual_server_id}/mcp",
		AdminPrefix: "/api",
		Health:      true,
	}
}

// New builds a Server with the given dependencies and configuration.
func New(deps Deps, cfg Config) *Server {
	r := cfg.Router
	if r == nil {
		r = mux.NewRouter()
	}
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Recover(deps.Logger))

	// Optional Basic Auth: if Authorization header present, validate and set context
	if deps.BasicUsername != "" {
		r.Use(middlewares.BasicAuth(deps.BasicUsername, deps.BasicPassword, deps.AdminUserID))
	}

	if deps.JWTSecret != "" {
		r.Use(middlewares.Auth(deps.JWTSecret))
	}

	for _, m := range cfg.Middlewares {
		r.Use(m)
	}

	// Require auth for everything except health and auth endpoints and static root.
	r.Use(middlewares.RequireAuthExcept(
		"/live", "/ready", cfg.AdminPrefix+"/auth", cfg.MCPMount,
	))

	// Mount routes
	addAuthRoutes(r, deps, cfg)
	addMCPRoutes(r, deps, cfg)
	addAdminRoutes(r, deps, cfg)
	if cfg.Health {
		addHealthRoutes(r, cfg)
	}

	return &Server{Handler: r}
}
