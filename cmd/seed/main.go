// Command seed seeds catalog data (mcp_servers) from config/mcp_servers.json and exits.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	cfgpkg "github.com/ChiragChiranjib/mcp-proxy/internal/config"
	logpkg "github.com/ChiragChiranjib/mcp-proxy/internal/log"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	mrepo "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
	"gorm.io/gorm/clause"
)

type catalogServer struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Description string `json:"description"`
}

func main() {
	jsonPath := flag.String("file", filepath.Join("config", "mcp_servers.json"), "catalog json path")
	flag.Parse()

	cfg, err := cfgpkg.Load()
	if err != nil {
		panic(err)
	}
	logger := logpkg.New(logpkg.Options{Level: slog.LevelInfo})

	grepo, err := mrepo.NewFromConfig(cfg.DB)
	if err != nil {
		logger.Error("gorm init", "error", err)
		os.Exit(1)
	}

	f, err := os.Open(*jsonPath)
	if err != nil {
		logger.Error("open file", "error", err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()
	content, err := io.ReadAll(f)
	if err != nil {
		logger.Error("read file", "error", err)
		os.Exit(1)
	}
	var items []catalogServer
	if err := json.Unmarshal(content, &items); err != nil {
		logger.Error("json", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	inserted := 0
	for _, it := range items {
		if it.Name == "" || it.URL == "" {
			continue
		}
		id := idgen.NewID()
		rec := m.MCPServer{ID: id, Name: it.Name, URL: it.URL, Description: it.Description}
		if err := grepo.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).Create(&rec).Error; err != nil {
			logger.Error("insert", "name", it.Name, "error", err)
			os.Exit(1)
		}
		inserted++
	}
	logger.Info("seeded mcp_servers", "count", inserted, "file", filepath.Base(*jsonPath))
}
