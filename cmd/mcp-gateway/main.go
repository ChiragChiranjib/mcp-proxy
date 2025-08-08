// Command mcp-gateway starts the Global MCP Gateway HTTP server.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	cfgpkg "github.com/ChiragChiranjib/mcp-proxy/internal/config"
	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	logpkg "github.com/ChiragChiranjib/mcp-proxy/internal/log"
	mrepo "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/catalog"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/tool"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/user"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/virtualmcp"
	mcpserver "github.com/ChiragChiranjib/mcp-proxy/internal/server"
)

func main() {
	// server flags only (no seeding here)
	flag.Parse()

	cfg, err := cfgpkg.Load()
	if err != nil {
		panic(err)
	}

	logger := logpkg.New(logpkg.Options{Level: slog.LevelInfo})

	// Initialize GORM repo for services using config
	grepo, err := mrepo.NewFromConfig(cfg.DB)
	if err != nil {
		logger.Error("gorm init", "error", err)
		os.Exit(1)
	}
	toolSvc := tool.NewService(
		tool.WithLogger(logger),
		tool.WithRepo(grepo),
	)

	hubSvc := mcphub.NewService(
		mcphub.WithLogger(logger),
		mcphub.WithRepo(grepo),
	)

	virtualSvc := virtualmcp.NewService(
		virtualmcp.WithLogger(logger),
		virtualmcp.WithRepo(grepo),
	)

	catalogSvc := catalog.NewService(
		catalog.WithLogger(logger),
		catalog.WithRepo(grepo),
	)
	userSvc := user.NewService(
		user.WithLogger(logger),
		user.WithRepo(grepo),
	)
	// Build AESEncrypter if key present
	var encr *encryptor.AESEncrypter
	if cfg.Security.AESKey != "" {
		if e, err := encryptor.NewAESEncrypter(cfg.Security.AESKey); err == nil {
			encr = e
		}
	}
	server := mcpserver.New(
		mcpserver.DefaultConfig(),
		mcpserver.WithLogger(logger),
		mcpserver.WithTools(toolSvc),
		mcpserver.WithHubs(hubSvc),
		mcpserver.WithVirtual(virtualSvc),
		mcpserver.WithCatalog(catalogSvc),
		mcpserver.WithUserService(userSvc),
		mcpserver.WithEncrypter(encr),
		mcpserver.WithAppConfig(cfg),
	)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", cfg.Server.Port), Handler: server.Handler}

	go func() {
		logger.Info("http server starting", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("http server", "error", err)
			os.Exit(1)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
