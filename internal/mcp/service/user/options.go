package user

import (
	"database/sql"
	"log/slog"
)

type Option interface{ apply(*Service) }
type withLogger struct{ l *slog.Logger }

func (o withLogger) apply(s *Service)  { s.logger = o.l }
func WithLogger(l *slog.Logger) Option { return withLogger{l} }

func NewService(db *sql.DB, opts ...Option) *Service {
	s := &Service{db: db}
	for _, o := range opts {
		o.apply(s)
	}
	return s
}
