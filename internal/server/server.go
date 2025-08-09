// Package server wires dependencies and builds the HTTP handler container.
package server

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	cfgpkg "github.com/ChiragChiranjib/mcp-proxy/internal/config"
	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/catalog"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	orchestrator "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub_orchestrator"
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
	Logger       *slog.Logger
	Tools        *tool.Service
	Hubs         *mcphub.Service
	Virtual      *virtualmcp.Service
	Catalog      *catalog.Service
	Encrypter    *encryptor.AESEncrypter
	UserService  *usersvc.Service
	Orchestrator *orchestrator.Orchestrator
	// Full app config for auth and other settings
	AppConfig *cfgpkg.Config
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

// Option configures dependencies for the server.
type Option func(*Deps)

func WithLogger(l *slog.Logger) Option               { return func(d *Deps) { d.Logger = l } }
func WithTools(s *tool.Service) Option               { return func(d *Deps) { d.Tools = s } }
func WithHubs(s *mcphub.Service) Option              { return func(d *Deps) { d.Hubs = s } }
func WithVirtual(s *virtualmcp.Service) Option       { return func(d *Deps) { d.Virtual = s } }
func WithCatalog(s *catalog.Service) Option          { return func(d *Deps) { d.Catalog = s } }
func WithUserService(s *usersvc.Service) Option      { return func(d *Deps) { d.UserService = s } }
func WithEncrypter(e *encryptor.AESEncrypter) Option { return func(d *Deps) { d.Encrypter = e } }
func WithAppConfig(c *cfgpkg.Config) Option          { return func(d *Deps) { d.AppConfig = c } }
func WithOrchestrator(o *orchestrator.Orchestrator) Option {
	return func(d *Deps) { d.Orchestrator = o }
}

// New builds a Server with the given configuration and dependency options.
func New(cfg Config, opts ...Option) *Server {
	var deps Deps
	for _, o := range opts {
		o(&deps)
	}
	r := cfg.Router
	if r == nil {
		r = mux.NewRouter()
	}
	r.Use(middlewares.RequestID())
	r.Use(middlewares.Recover(deps.Logger))

	if deps.AppConfig != nil && deps.AppConfig.Security.BasicUsername != "" {
		r.Use(middlewares.BasicAuth(
			deps.AppConfig.Security.BasicUsername,
			deps.AppConfig.Security.BasicPassword,
			deps.AppConfig.Security.AdminUserID,
			deps.Logger,
		))
	}

	if deps.AppConfig != nil && deps.AppConfig.Security.JWTSecret != "" {
		r.Use(middlewares.Auth(deps.AppConfig.Security.JWTSecret, deps.Logger))
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
