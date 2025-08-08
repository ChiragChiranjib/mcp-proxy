// Package storage provides database connectivity helpers for internal use.
package storage

import (
	"database/sql"
	"fmt"
	"time"

	// mysql driver
	_ "github.com/go-sql-driver/mysql"
)

// Config holds MySQL pool configuration.
type Config struct {
	DSN                    string
	MaxOpenConns           int
	MaxIdleConns           int
	ConnMaxIdleSeconds     int
	ConnMaxLifetimeSeconds int
}

// NewMySQL returns a configured *sql.DB connection pool.
func NewMySQL(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	if cfg.MaxOpenConns > 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxIdleSeconds > 0 {
		db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleSeconds) * time.Second)
	}
	if cfg.ConnMaxLifetimeSeconds > 0 {
		db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeSeconds) * time.Second)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}
