-- name: AddHubServer :exec
INSERT INTO mcp_hub_servers (id, user_id, mcp_server_id, status, transport, capabilities, auth_type, auth_value)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);

-- name: UpdateHubServerStatus :exec
UPDATE mcp_hub_servers SET status = ? WHERE id = ?;

-- name: DeleteHubServer :exec
DELETE FROM mcp_hub_servers WHERE id = ?;

-- name: GetHubServer :one
SELECT id, user_id, mcp_server_id, status, transport, capabilities, auth_type, auth_value, created_at, updated_at
FROM mcp_hub_servers WHERE id = ? LIMIT 1;

-- name: ListUserHubServers :many
SELECT id, user_id, mcp_server_id, status, transport, capabilities, auth_type, auth_value, created_at, updated_at
FROM mcp_hub_servers WHERE user_id = ? ORDER BY created_at DESC;

-- name: GetHubServerWithURL :one
SELECT hs.id,
       hs.user_id,
       hs.mcp_server_id,
       hs.status,
       hs.transport,
       hs.capabilities,
       hs.auth_type,
       hs.auth_value,
       s.url AS server_url,
       s.name AS server_name
FROM mcp_hub_servers hs
JOIN mcp_servers s ON s.id = hs.mcp_server_id
WHERE hs.id = ?
LIMIT 1;
