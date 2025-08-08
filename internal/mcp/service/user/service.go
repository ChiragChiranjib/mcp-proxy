// Package user provides minimal user management for SSO.
package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	sqldb "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo/db"
)

type Service struct {
	q      *sqldb.Queries
	db     *sql.DB
	logger *slog.Logger
}

// GetOrCreateByEmail returns an existing user id for the given email (stored as username),
// or creates a new user with role USER.
func (s *Service) GetOrCreateByEmail(ctx context.Context, email string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username=? LIMIT 1", email).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	uid := idgen.NewID()
	if _, err := s.db.ExecContext(ctx, "INSERT INTO users (id, username, role) VALUES (?, ?, ?)", uid, email, "USER"); err != nil {
		// possible race: try re-select
		_ = s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username=? LIMIT 1", email).Scan(&id)
		if id != "" {
			return id, nil
		}
		return "", err
	}
	return uid, nil
}
