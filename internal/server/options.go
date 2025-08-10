package server

import (
	"log/slog"

	cfgpkg "github.com/ChiragChiranjib/mcp-proxy/internal/config"
	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/catalog"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	orchestrator "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub_orchestrator"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/tool"
	usersvc "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/user"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/virtualmcp"
)

// Option configures dependencies for the server.
type Option func(*Deps)

// WithLogger ...
func WithLogger(l *slog.Logger) Option {
	return func(d *Deps) {
		d.Logger = l
	}
}

// WithTools ...
func WithTools(s *tool.Service) Option {
	return func(d *Deps) {
		d.Tools = s
	}
}

// WithHubs ...
func WithHubs(s *mcphub.Service) Option {
	return func(d *Deps) {
		d.Hubs = s
	}
}

// WithVirtual ...
func WithVirtual(s *virtualmcp.Service) Option {
	return func(d *Deps) {
		d.Virtual = s

	}
}

// WithCatalog ...
func WithCatalog(s *catalog.Service) Option {
	return func(d *Deps) {
		d.Catalog = s
	}
}

// WithUserService ...
func WithUserService(s *usersvc.Service) Option {
	return func(d *Deps) {
		d.UserService = s
	}
}

// WithEncrypter ...
func WithEncrypter(e *encryptor.AESEncrypter) Option {
	return func(d *Deps) {
		d.Encrypter = e
	}
}

// WithAppConfig ...
func WithAppConfig(c *cfgpkg.Config) Option {
	return func(d *Deps) {
		d.AppConfig = c
	}
}

// WithOrchestrator ...
func WithOrchestrator(o *orchestrator.Orchestrator) Option {
	return func(d *Deps) { d.Orchestrator = o }
}
