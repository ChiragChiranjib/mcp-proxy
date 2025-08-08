package repo

import (
    "context"

    m "github.com/ChiragChiranjib/mcp-proxy/internal/models"
)

// GetUserByUsername returns the full user record by username (email).
func (r *Repo) GetUserByUsername(ctx context.Context, username string) (m.User, error) {
    var u m.User
    err := r.WithContext(ctx).Where("username = ?", username).Take(&u).Error
    return u, err
}

// FindUserIDByUsername returns only the id for a given username (email).
func (r *Repo) FindUserIDByUsername(ctx context.Context, username string) (string, error) {
    var u m.User
    if err := r.WithContext(ctx).Select("id").Where("username = ?", username).Take(&u).Error; err != nil {
        return "", err
    }
    return u.ID, nil
}

// CreateUser inserts a new user.
func (r *Repo) CreateUser(ctx context.Context, u m.User) error {
    return r.WithContext(ctx).Create(&u).Error
}

