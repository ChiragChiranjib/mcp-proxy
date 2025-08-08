// Command migrate applies SQL migrations from the migrations/ directory.
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"

	cfgpkg "github.com/ChiragChiranjib/mcp-proxy/internal/config"
	logpkg "github.com/ChiragChiranjib/mcp-proxy/internal/log"
	"github.com/ChiragChiranjib/mcp-proxy/internal/storage"
)

func main() {
	logger := logpkg.New(logpkg.Options{Level: slog.LevelInfo})

	cfg, err := cfgpkg.Load()
	if err != nil {
		logger.Error("load config", "error", err)
		os.Exit(1)
	}

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
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true", cfg.DB.Username, cfg.DB.Password, host, port, cfg.DB.Name)
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
	defer func() { _ = db.Close() }()

	if err := ensureMigrationsTable(db); err != nil {
		logger.Error("ensure migrations table", "error", err)
		os.Exit(1)
	}

	applied, err := readApplied(db)
	if err != nil {
		logger.Error("read applied", "error", err)
		os.Exit(1)
	}

	files, err := collectMigrationFiles("./migrations")
	if err != nil {
		logger.Error("collect migrations", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	for _, f := range files {
		name := filepath.Base(f)
		if applied[name] {
			continue
		}
		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			logger.Error("read migration", "file", name, "error", err)
			os.Exit(1)
		}
		up := extractUp(string(sqlBytes))
		if err := execStatements(ctx, db, up); err != nil {
			logger.Error("apply migration", "file", name, "error", err)
			os.Exit(1)
		}
		if err := recordApplied(db, name); err != nil {
			logger.Error("record migration", "file", name, "error", err)
			os.Exit(1)
		}
		logger.Info("applied migration", "file", name)
	}
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  version VARCHAR(255) PRIMARY KEY,
  applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	return err
}

func readApplied(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		if isNoSuchTable(err) {
			return map[string]bool{}, nil
		}
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	m := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		m[v] = true
	}
	return m, rows.Err()
}

func collectMigrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) == ".sql" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func execStatements(ctx context.Context, db *sql.DB, script string) error {
	// For MySQL, recommend DSN with multiStatements=true. If not present, this will fail
	// for multi-statement files. Keep scripts single-statement or enable multiStatements.
	_, err := db.ExecContext(ctx, script)
	return err
}

// extractUp returns the content between goose Up and Down markers if present, else the script as-is.
func extractUp(script string) string {
	lower := strings.ToLower(script)
	upIdx := strings.Index(lower, "-- +goose up")
	if upIdx == -1 {
		return script
	}
	downIdx := strings.Index(lower, "-- +goose down")
	if downIdx == -1 {
		return script[upIdx+len("-- +goose up"):]
	}
	return script[upIdx+len("-- +goose up") : downIdx]
}

func recordApplied(db *sql.DB, version string) error {
	_, err := db.Exec("INSERT INTO schema_migrations(version) VALUES(?)", version)
	return err
}

func isNoSuchTable(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
