-- name: CreateVirtualServer :exec
INSERT INTO mcp_virtual_servers (id, user_id, status) VALUES (?, ?, ?);

-- name: UpdateVirtualServerStatus :exec
UPDATE mcp_virtual_servers SET status = ? WHERE id = ?;

-- name: ListVirtualServersForUser :many
SELECT id, user_id, status, created_at, updated_at FROM mcp_virtual_servers WHERE user_id = ? ORDER BY created_at DESC;

-- name: ReplaceVirtualServerTools :exec
DELETE FROM tools_virtual_servers WHERE mcp_virtual_server_id = ?;

-- name: AddVirtualServerTool :exec
INSERT INTO tools_virtual_servers (mcp_virtual_server_id, tool_id) VALUES (?, ?);

-- name: ListToolsForVirtualServer :many
SELECT t.id, t.user_id, t.original_name, t.modified_name, t.mcp_hub_server_id, t.input_schema, t.annotations, t.status, t.created_at, t.updated_at
FROM tools_virtual_servers v
JOIN mcp_tools t ON t.id = v.tool_id
WHERE v.mcp_virtual_server_id = ? AND t.status = 'ACTIVE';

-- name: DeleteVirtualServer :exec
DELETE FROM mcp_virtual_servers WHERE id = ?;
