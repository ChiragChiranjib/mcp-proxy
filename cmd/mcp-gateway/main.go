// Command mcp-gateway starts the Global MCP Gateway HTTP server.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	cfgpkg "github.com/ChiragChiranjib/mcp-proxy/internal/config"
	"github.com/ChiragChiranjib/mcp-proxy/internal/encryptor"
	logpkg "github.com/ChiragChiranjib/mcp-proxy/internal/log"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/catalog"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/mcphub"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/tool"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/user"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/virtualmcp"
	mcpserver "github.com/ChiragChiranjib/mcp-proxy/internal/server"
	"github.com/ChiragChiranjib/mcp-proxy/internal/storage"
)

type catalogServer struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

func seedMCPCatalog(ctx context.Context, db *sql.DB, jsonPath string, logger *slog.Logger) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	var items []catalogServer
	if err := json.Unmarshal(content, &items); err != nil {
		return err
	}
	// Simple upsert by unique name
	const q = `INSERT INTO mcp_servers (id, name, url, description) VALUES (?, ?, ?, ?) ON DUPLICATE KEY UPDATE id = id`
	inserted := 0
	for _, it := range items {
		if it.Name == "" || it.URL == "" {
			continue
		}
		id := idgen.NewID()
		if _, err := db.ExecContext(ctx, q, id, it.Name, it.URL, it.Description); err != nil {
			return err
		}
		inserted++
	}
	logger.Info("seeded mcp_servers", "count", inserted, "file", filepath.Base(jsonPath))
	return nil
}

func main() {
	var seedCatalog bool
	flag.BoolVar(&seedCatalog, "seed-mcp-servers", false, "seed mcp_servers from config/mcp_servers.json and exit")
	flag.Parse()

	cfg, err := cfgpkg.Load()
	if err != nil {
		panic(err)
	}

	logger := logpkg.New(logpkg.Options{Level: slog.LevelInfo})

	dsn := cfg.DB.DSN
	if dsn == "" {
		host := cfg.DB.Host
		if host == "" {
			host = "127.0.0.1"
		}
		port := cfg.DB.Port
		if port == 0 {
			port = 3306
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
			cfg.DB.Username, cfg.DB.Password, host, port, cfg.DB.Name,
		)
	}
	db, err := storage.NewMySQL(storage.Config{
		DSN:                    dsn,
		MaxOpenConns:           cfg.DB.MaxOpenConns,
		MaxIdleConns:           cfg.DB.MaxIdleConns,
		ConnMaxIdleSeconds:     cfg.DB.ConnMaxIdleSeconds,
		ConnMaxLifetimeSeconds: cfg.DB.ConnMaxLifetimeSeconds,
	})
	if err != nil {
		logger.Error("db init", "error", err)
		os.Exit(1)
	}
	// Optional one-shot seeding
	if seedCatalog {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := seedMCPCatalog(ctx, db, filepath.Join("config", "mcp_servers.json"), logger); err != nil {
			logger.Error("seed catalog", "error", err)
			os.Exit(1)
		}
		logger.Info("seed completed; exiting as requested by flag")
		return
	}

	toolSvc := tool.NewService(db, tool.WithLogger(logger))
	hubSvc := mcphub.NewService(db, mcphub.WithLogger(logger))
	virtualSvc := virtualmcp.NewService(db, virtualmcp.WithLogger(logger))
	catalogSvc := catalog.NewService(db, catalog.WithLogger(logger))
	userSvc := user.NewService(db, user.WithLogger(logger))
	// Build AESEncrypter if key present
	var encr *encryptor.AESEncrypter
	if cfg.Security.AESKey != "" {
		if e, err := encryptor.NewAESEncrypter(cfg.Security.AESKey); err == nil {
			encr = e
		}
	}
	server := mcpserver.New(mcpserver.Deps{
		Logger:         logger,
		Tools:          toolSvc,
		Hubs:           hubSvc,
		Virtual:        virtualSvc,
		Catalog:        catalogSvc,
		Encrypter:      encr,
		GoogleClientID: cfg.Google.ClientID,
		JWTSecret:      cfg.Security.JWTSecret,
		UserService:    userSvc,
		BasicUsername:  cfg.Security.BasicUsername,
		BasicPassword:  cfg.Security.BasicPassword,
		AdminUserID:    cfg.Security.AdminUserID,
	}, mcpserver.DefaultConfig())

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
