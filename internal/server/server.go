// Package server wires dependencies and builds the HTTP handler container.
package server

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/tool"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/virtualmcp"
	"github.com/ChiragChiranjib/mcp-proxy/internal/middlewares"
)

// Server holds the final HTTP handler for the app.
type Server struct {
	Handler http.Handler
}

// Deps enumerates dependencies required to assemble the server.
type Deps struct {
	Logger    *slog.Logger
	Tools     *tool.Service
	Hubs      *mcphub.Service
	Virtual   *virtualmcp.Service
	Encrypter *encryptor.AESEncrypter
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
	for _, m := range cfg.Middlewares {
		r.Use(m)
	}

	// Mount routes
	addMCPRoutes(r, deps, cfg)
	addAdminRoutes(r, deps, cfg)
	if cfg.Health {
		addHealthRoutes(r, cfg)
	}

	return &Server{Handler: r}
}
