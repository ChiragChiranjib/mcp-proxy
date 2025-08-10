package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { goose.AddMigrationContext(upCreateMCPVirtualServers, downCreateMCPVirtualServers) }

func upCreateMCPVirtualServers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS mcp_virtual_servers (
  id CHAR(22) PRIMARY KEY,
  user_id CHAR(22) NOT NULL,
  name VARCHAR(255) NOT NULL,
  status VARCHAR(30) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  INDEX idx_vs_user (user_id),
  INDEX idx_vs_status (status),
  INDEX idx_vs_user_status (user_id, status),
  INDEX idx_vs_user_name (user_id, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`)
	return err
}

func downCreateMCPVirtualServers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS mcp_virtual_servers;`)
	return err
}
