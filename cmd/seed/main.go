// Package main seeds catalog data and users from JSON files.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
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

type seedUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func main() {
	serversPath := flag.String("servers", filepath.Join("cmd", "seed", "data", "mcp_servers.json"), "catalog servers json path")
	usersPath := flag.String("users", filepath.Join("cmd", "seed", "data", "users.json"), "users json path")
	only := flag.String("only", "", "seed only: servers|users (default both)")
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if *only == "" || *only == "servers" {
		if err := seedServers(ctx, grepo, *serversPath, logger); err != nil {
			logger.Error("seed servers", "error", err)
			os.Exit(1)
		}
	}

	if *only == "" || *only == "users" {
		if err := seedUsers(ctx, grepo, *usersPath, logger); err != nil {
			logger.Error("seed users", "error", err)
			os.Exit(1)
		}
	}
}

func seedServers(ctx context.Context, grepo *mrepo.Repo, jsonPath string, logger *slog.Logger) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()
	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	var items []catalogServer
	if err := json.Unmarshal(content, &items); err != nil {
		return fmt.Errorf("json: %w", err)
	}
	inserted := 0
	for _, it := range items {
		if it.Name == "" || it.URL == "" {
			continue
		}
		id := idgen.NewID()
		rec := m.MCPServer{ID: id, Name: it.Name, URL: it.URL, Description: it.Description}
		if err := grepo.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "name"}}, DoNothing: true}).Create(&rec).Error; err != nil {
			return fmt.Errorf("insert server %s: %w", it.Name, err)
		}
		inserted++
	}
	logger.Info("seeded mcp_servers", "count", inserted, "file", filepath.Base(jsonPath))
	return nil
}

func seedUsers(ctx context.Context, grepo *mrepo.Repo, jsonPath string, logger *slog.Logger) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("users.json not found; skipping", "path", jsonPath)
			return nil
		}
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()
	content, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}
	var items []seedUser
	if err := json.Unmarshal(content, &items); err != nil {
		return fmt.Errorf("json: %w", err)
	}
	inserted := 0
	for _, it := range items {
		username := it.Username
		if username == "" {
			continue
		}
		// generate id if missing
		id := it.ID
		if id == "" {
			id = idgen.NewID()
		}
		role := it.Role
		if role == "" {
			role = "USER"
		}
		rec := m.User{ID: id, Username: username, Role: role}
		// upsert by username
		if err := grepo.WithContext(ctx).Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "username"}}, DoUpdates: clause.Assignments(map[string]any{"role": role})}).Create(&rec).Error; err != nil {
			return fmt.Errorf("insert user %s: %w", username, err)
		}
		inserted++
	}
	logger.Info("seeded users", "count", inserted, "file", filepath.Base(jsonPath))
	return nil
}
