// Package catalog provides access to the MCP servers catalog.
package catalog

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	sqldb "github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo/db"
	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/service/types"
)

// Service exposes catalog operations.
type Service struct {
	q       *sqldb.Queries
	logger  *slog.Logger
	timeout time.Duration
}

// Option configures the Service.
type Option interface{ apply(*Service) }

type withLogger struct{ l *slog.Logger }

func (o withLogger) apply(s *Service) { s.logger = o.l }

// WithLogger sets the logger for the service.
func WithLogger(l *slog.Logger) Option { return withLogger{l} }

// NewService creates a new catalog Service.
func NewService(db *sql.DB, opts ...Option) *Service {
	s := &Service{q: sqldb.New(db)}
	for _, o := range opts {
		o.apply(s)
	}
	return s
}

func (s *Service) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if s.timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, s.timeout)
}

// List returns all catalog servers ordered by name.
func (s *Service) List(ctx context.Context) ([]types.CatalogServer, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	rows, err := s.q.ListCatalogServers(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]types.CatalogServer, 0, len(rows))
	for _, r := range rows {
		desc := ""
		if r.Description.Valid {
			desc = r.Description.String
		}
		out = append(out, types.CatalogServer{ID: r.ID, Name: r.Name, URL: r.Url, Description: desc})
	}
	return out, nil
}
