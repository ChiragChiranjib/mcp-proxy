package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() { goose.AddMigrationContext(upCreateToolsVirtualServers, downCreateToolsVirtualServers) }

func upCreateToolsVirtualServers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS tools_virtual_servers (
  mcp_virtual_server_id CHAR(22) NOT NULL,
  tool_id CHAR(22) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (mcp_virtual_server_id, tool_id),
  CONSTRAINT fk_vs_tool_vs FOREIGN KEY (mcp_virtual_server_id) REFERENCES mcp_virtual_servers(id) ON DELETE CASCADE,
  INDEX idx_vs_tool_vs (mcp_virtual_server_id),
  INDEX idx_vs_tool_tool (tool_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`)
	return err
}

func downCreateToolsVirtualServers(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE IF EXISTS tools_virtual_servers;`)
	return err
}
