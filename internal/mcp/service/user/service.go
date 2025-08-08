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
	var u m.User
	err := s.repo.WithContext(ctx).Where("username = ?", email).Take(&u).Error
	if err == nil && u.ID != "" {
		return u.ID, nil
	}
	// create
	uid := idgen.NewID()
	u = m.User{ID: uid, Username: email, Role: "USER"}
	if err := s.repo.WithContext(ctx).Create(&u).Error; err != nil {
		// race-safe reselect
		var again m.User
		if e2 := s.repo.WithContext(ctx).Where("username = ?", email).Take(&again).Error; e2 == nil && again.ID != "" {
			return again.ID, nil
		}
		return "", err
	}
	return uid, nil
}
