package service

import (
	"pkg/token"
	"reg/hasher"
	"reg/repo"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo      *repo.Repositories
	hasher    hasher.IHasher
	generator token.ITokenGenerator
}

func New(db *pgxpool.Pool, hasher hasher.IHasher) *Service {
	return &Service{repo: repo.NewRepositories(db), hasher: hasher}
}
