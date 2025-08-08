package user

import (
	"log/slog"

	"github.com/ChiragChiranjib/mcp-proxy/internal/mcp/repo"
)

type Option func(*Service)

func WithLogger(l *slog.Logger) Option { return func(s *Service) { s.logger = l } }

func WithRepo(r *repo.Repo) Option { return func(s *Service) { s.repo = r } }

func NewService(opts ...Option) *Service {
	s := &Service{}
	for _, o := range opts {
		o(s)
	}
	return s
}
