-- name: CreateUser :exec
INSERT INTO users (id, username, role) VALUES (?, ?, ?);

-- name: GetUser :one
SELECT id, username, role, created_at, updated_at FROM users WHERE id = ? LIMIT 1;
