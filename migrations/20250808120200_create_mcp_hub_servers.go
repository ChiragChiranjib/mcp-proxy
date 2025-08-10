package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { goose.AddMigrationContext(upCreateMCPHubServers, downCreateMCPHubServers) }

func upCreateMCPHubServers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS mcp_hub_servers (
  id CHAR(22) PRIMARY KEY,
  user_id CHAR(22) NOT NULL,
  mcp_server_id CHAR(22) NOT NULL,
  status VARCHAR(30) NOT NULL,
  transport VARCHAR(30) NOT NULL,
  capabilities JSON,
  auth_type VARCHAR(30) NOT NULL,
  auth_value JSON,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_hub_mcp_srv FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE,
  UNIQUE KEY uq_user_server (user_id, mcp_server_id),
  INDEX idx_hub_user (user_id),
  INDEX idx_hub_mcp (mcp_server_id),
  INDEX idx_hub_user_status (user_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`)
	return err
}

func downCreateMCPHubServers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS mcp_hub_servers;`)
	return err
}
