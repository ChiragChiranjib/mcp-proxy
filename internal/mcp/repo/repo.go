package repo

import (
	"fmt"

	"github.com/ChiragChiranjib/mcp-proxy/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Repo is a thin wrapper over gorm.DB to centralize data access.
type Repo struct {
	*gorm.DB
}

// New creates a Repo from DSN.
func New(dsn string) (*Repo, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Repo{DB: db}, nil
}

// NewFromConfig builds the DSN from app config and opens a Repo.
func NewFromConfig(dbCfg config.DatabaseConfig) (*Repo, error) {
	dsn := dbCfg.DSN
	if dsn == "" {
		host := dbCfg.Host
		if host == "" {
			host = "127.0.0.1"
		}
		port := dbCfg.Port
		if port == 0 {
			port = 3306
		}
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&multiStatements=true",
			dbCfg.Username, dbCfg.Password, host, port, dbCfg.Name,
		)
	}
	return New(dsn)
}

// Transaction helper that accepts a function with *Repo operating on same DB
func (r *Repo) Transaction(fn func(tx *Repo) error) error {
	return r.DB.Transaction(func(txdb *gorm.DB) error {
		return fn(&Repo{DB: txdb})
	})
}