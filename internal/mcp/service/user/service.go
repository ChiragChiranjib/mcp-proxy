// Package user provides minimal user management for SSO.
package user

import (
	"context"
	"log/slog"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

type Service struct {
	repo   *repo.Repo
	logger *slog.Logger
}

// GetOrCreateByEmail returns an existing user id for the given email (stored as username),
// or creates a new user with role USER.
func (s *Service) GetOrCreateByEmail(ctx context.Context, email string) (string, error) {
	if id, err := s.repo.FindUserIDByUsername(ctx, email); err == nil && id != "" {
		return id, nil
	}
	// create
	uid := idgen.NewID()
	u := m.User{ID: uid, Username: email, Role: "USER"}
	if err := s.repo.CreateUser(ctx, u); err != nil {
		if id, err2 := s.repo.FindUserIDByUsername(ctx, email); err2 == nil && id != "" {
			return id, nil
		}
		return "", err
	}
	return uid, nil
}
