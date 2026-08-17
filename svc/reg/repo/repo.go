package repo

import (
	"context"
	userv1 "proto/out/user/v1"
	"reg/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Repositories struct {
	user    *userRepo
	refresh *refreshRepo
}

func NewRepositories(pool *pgx.Conn) *Repositories {
	queries := db.New(pool)
	user := &userRepo{db: pool, queries: queries}
	refresh := &refreshRepo{db: pool, queries: queries}
	return &Repositories{user: user, refresh: refresh}
}

func (r *Repositories) RegisterUser(ctx context.Context,
	params *userv1.RegisterUserRequest,
	passwordHash string) (CreateResponse, error) {
	arg := db.CreateUserParams{Email: params.GetEmail(), Name: params.GetName(), PasswordHash: passwordHash}
	res, err := r.user.RegisterUser(ctx, arg)
	if err != nil {
		return CreateResponse{}, err
	}
	return res, nil
}

func (r *Repositories) LoginUser(ctx context.Context,
	params *userv1.LoginUserRequest,
) (db.FetchPasswordRow, error) {
	res, err := r.user.LoginUser(ctx, params.GetEmail())
	if err != nil {
		return db.FetchPasswordRow{}, err
	}
	return res, nil
}

func (r *Repositories) CreateToken(ctx context.Context, id uuid.UUID, tokenHash string) error {
	err := r.refresh.CreateToken(ctx, id, tokenHash)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repositories) DeleteToken(ctx context.Context, tokenHash string) error {
	err := r.refresh.DeleteToken(ctx, tokenHash)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repositories) GetProfile(ctx context.Context, id uuid.UUID) (FetchProfileResponse, error) {
	res, err := r.user.FetchProfile(ctx, id)
	if err != nil {
		return FetchProfileResponse{}, err
	}
	return res, nil
}
