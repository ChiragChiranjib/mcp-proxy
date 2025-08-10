// Package user provides minimal user management for SSO.
package user

import (
	"context"
	"log/slog"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/idgen"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// Service ...
type Service struct {
	repo   *repo.Repo
	logger *slog.Logger
}

// FetchOrCreateByUsername returns an existing user id for the given email
// (stored as username), or creates a new user with role USER.
func (s *Service) FetchOrCreateByUsername(
	ctx context.Context, username string) (*m.User, error) {
	user, err := s.repo.FindUserByUsername(ctx, username)
	if err == nil && user.ID != "" {
		return user, nil
	}

	// Create a new user
	u := &m.User{
		ID:       idgen.NewID(),
		Username: username,
		Role:     string(m.RoleUser),
	}
	if err = s.repo.CreateUser(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

// FindUserByUserName ...
func (s *Service) FindUserByUserName(
	ctx context.Context,
	username string) (*m.User, error) {
	user, err := s.repo.FindUserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return user, nil
}
