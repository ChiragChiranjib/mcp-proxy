package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { goose.AddMigrationContext(upCreateMCPTools, downCreateMCPTools) }

func upCreateMCPTools(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS mcp_tools (
  id CHAR(22) PRIMARY KEY,
  user_id CHAR(22),
  mcp_server_id CHAR(22) NOT NULL,
  mcp_hub_server_id CHAR(22),
  original_name VARCHAR(255) NOT NULL,
  modified_name VARCHAR(255) NOT NULL,
  description TEXT,
  input_schema JSON,
  annotations JSON,
  status VARCHAR(30) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  CONSTRAINT fk_tools_server FOREIGN KEY (mcp_server_id) REFERENCES mcp_servers(id) ON DELETE CASCADE,
  CONSTRAINT fk_tools_hub_server FOREIGN KEY (mcp_hub_server_id) REFERENCES mcp_hub_servers(id) ON DELETE CASCADE,
  UNIQUE KEY uq_server_user_tool (mcp_server_id, user_id, original_name),
  INDEX idx_tools_user (user_id),
  INDEX idx_tools_server (mcp_server_id),
  INDEX idx_tools_hub_server (mcp_hub_server_id),
  INDEX idx_tools_status (status),
  INDEX idx_tools_server_user (mcp_server_id, user_id),
  INDEX idx_tools_user_status (user_id, status),
  INDEX idx_tools_server_status (mcp_server_id, status),
  INDEX idx_tools_hub_server_status (mcp_hub_server_id, status),
  INDEX idx_tools_id_status (id, status),
  INDEX idx_tools_names (modified_name, original_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`)
	return err
}

func downCreateMCPTools(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS mcp_tools;`)
	return err
}
