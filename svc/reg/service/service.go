package service

import (
	"context"
	"pkg/token"
	userv1 "proto/out/user/v1"
	"reg/hasher"
	"reg/repo"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	repo      *repo.Repositories
	hasher    hasher.IHasher
	generator token.ITokenGenerator
}

func New(db *pgxpool.Pool, hasher hasher.IHasher, generator token.ITokenGenerator) *Service {
	return &Service{repo: repo.NewRepositories(db), hasher: hasher, generator: generator}
}

func (s *Service) Register(ctx context.Context,
	params *userv1.RegisterUserRequest) (*userv1.RegisterUserResponse, error) {
	res := &userv1.RegisterUserResponse{}
	pwdHash, err := s.hasher.Hash(params.GetPassword())
	if err != nil {
		return nil, err
	}
	err = s.repo.WithinTx(ctx, func(ctx context.Context) error {
		_, err := s.repo.RegisterUser(ctx, params, pwdHash)
		if err != nil {
			return err
		}
		res.Response = "successfully registered"
		return nil
	})
	return res, nil
}
