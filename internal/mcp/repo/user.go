package repo

import (
	"context"

	m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// FindUserByUsername returns the full user record by username
func (r *Repo) FindUserByUsername(
	ctx context.Context, username string) (*m.User, error) {
	var u m.User
	if err := r.WithContext(ctx).
		Where("username = ?", username).
		Take(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

// CreateUser inserts a new user.
func (r *Repo) CreateUser(ctx context.Context, u *m.User) error {
	return r.WithContext(ctx).Create(u).Error
}
