// Package server wires dependencies and builds the HTTP handler container.
package server

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	cfgpkg "github.com/ChiragChiranjib/mcp-proxy/internal/config"
	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/catalog"
	catalogOrchestrator "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/catalog_orchestrator"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	mcphubOrchestrator "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub_orchestrator"
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
	Logger              *slog.Logger
	Tools               *tool.Service
	Hubs                *mcphub.Service
	Virtual             *virtualmcp.Service
	Catalog             *catalog.Service
	Encrypter           *encryptor.AESEncrypter
	UserService         *usersvc.Service
	McphubOrchestrator  *mcphubOrchestrator.Orchestrator
	CatalogOrchestrator *catalogOrchestrator.Orchestrator
	AppConfig           *cfgpkg.Config
}

// Config holds HTTP wiring configuration.
type Config struct {
	Middlewares []func(http.Handler) http.Handler
	MCPMount    string
	AdminPrefix string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		MCPMount:    "/servers/{virtual_server_id}/mcp",
		AdminPrefix: "/api",
	}
}

// New builds a Server with the given configuration and dependency options.
func New(cfg Config, opts ...Option) *Server {
	var deps Deps
	for _, o := range opts {
		o(&deps)
	}

	r := mux.NewRouter()
	r.Use(
		middlewares.Recover(deps.Logger),
		middlewares.Tag(),
		middlewares.Auth(
			deps.Logger,
			deps.AppConfig.Security.JWTSecret,
		),
		middlewares.BasicAuth(
			deps.Logger,
			deps.AppConfig.Security,
			deps.UserService,
		),
	)

	for _, m := range cfg.Middlewares {
		r.Use(m)
	}

	// Mount routes
	addAuthRoutes(r, deps, cfg)
	addMCPRoutes(r, deps, cfg)
	addAdminRoutes(r, deps, cfg)
	addHealthRoutes(r, cfg)

	return &Server{Handler: r}
}
