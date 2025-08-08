-- name: ListCatalogServers :many
SELECT id, name, url, description, created_at, updated_at FROM mcp_servers ORDER BY name;

-- name: GetMcpServer :one
SELECT id, name, url, description, created_at, updated_at FROM mcp_servers WHERE id = ? LIMIT 1;
