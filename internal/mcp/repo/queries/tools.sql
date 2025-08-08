-- name: UpsertTool :exec
INSERT INTO mcp_tools (id, user_id, original_name, modified_name, mcp_hub_server_id, input_schema, annotations, status)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  input_schema = VALUES(input_schema),
  annotations = VALUES(annotations),
  status = VALUES(status),
  updated_at = CURRENT_TIMESTAMP;

-- name: ListActiveToolsForHub :many
SELECT id, user_id, original_name, modified_name, mcp_hub_server_id, input_schema, annotations, status, created_at, updated_at
FROM mcp_tools WHERE mcp_hub_server_id = ? AND status = 'ACTIVE';

-- name: GetToolByModifiedName :one
SELECT id, user_id, original_name, modified_name, mcp_hub_server_id, input_schema, annotations, status, created_at, updated_at
FROM mcp_tools WHERE user_id = ? AND modified_name = ? LIMIT 1;

-- name: UpdateToolStatus :exec
UPDATE mcp_tools SET status = ? WHERE id = ?;

-- name: ListToolsForUserFiltered :many
SELECT id, user_id, original_name, modified_name, mcp_hub_server_id, input_schema, annotations, status, created_at, updated_at
FROM mcp_tools
WHERE user_id = ?
  AND (? = '' OR mcp_hub_server_id = ?)
  AND (? = '' OR status = ?)
  AND (? = '' OR (modified_name LIKE CONCAT('%', ?, '%') OR original_name LIKE CONCAT('%', ?, '%')))
ORDER BY modified_name;
