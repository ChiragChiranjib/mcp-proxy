package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { goose.AddMigrationContext(upCreateMCPServers, downCreateMCPServers) }

func upCreateMCPServers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS mcp_servers (
  id CHAR(22) PRIMARY KEY,
  name VARCHAR(255) NOT NULL UNIQUE,
  url VARCHAR(255) NOT NULL UNIQUE,
  description VARCHAR(255) DEFAULT '',
  capabilities JSON,
  transport VARCHAR(30) NOT NULL DEFAULT 'streamable-http',
  access_type VARCHAR(30) NOT NULL DEFAULT 'public',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_mcp_servers_name (name),
  INDEX idx_mcp_servers_access_type (access_type),
  INDEX idx_mcp_servers_access_name (access_type, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`)
	return err
}

func downCreateMCPServers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS mcp_servers;`)
	return err
}
