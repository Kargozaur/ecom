package repo

import (
	"context"
	userv1 "proto/out/user/v1"
	"reg/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type Repositories struct {
	pool    *pgxpool.Pool
	queries *db.Queries
	user    *userRepo
	refresh *refreshRepo
}

func NewRepositories(pool *pgxpool.Pool) *Repositories {
	queries := db.New(pool)
	user := &userRepo{}
	refresh := &refreshRepo{}
	return &Repositories{user: user, refresh: refresh, pool: pool, queries: queries}
}

func (r *Repositories) querier(ctx context.Context) *db.Queries {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return db.New(tx)
	}
	return r.queries
}

func (r *Repositories) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	ctx = context.WithValue(ctx, txKey{}, tx)
	if err := fn(ctx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (r *Repositories) RegisterUser(ctx context.Context,
	params *userv1.RegisterUserRequest,
	passwordHash string) (CreateResponse, error) {
	arg := db.CreateUserParams{Email: params.GetEmail(), Name: params.GetName(), PasswordHash: passwordHash}
	res, err := r.user.RegisterUser(ctx, r.querier(ctx), arg)
	if err != nil {
		return CreateResponse{}, err
	}
	return res, nil
}

func (r *Repositories) LoginUser(ctx context.Context,
	params *userv1.LoginUserRequest,
) (db.FetchPasswordRow, error) {
	res, err := r.user.LoginUser(ctx, r.querier(ctx), params.GetEmail())
	if err != nil {
		return db.FetchPasswordRow{}, err
	}
	return res, nil
}

func (r *Repositories) CreateToken(ctx context.Context, id uuid.UUID, tokenHash string) error {
	err := r.refresh.CreateToken(ctx, r.querier(ctx), id, tokenHash)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repositories) DeleteToken(ctx context.Context, tokenHash string) error {
	err := r.refresh.DeleteToken(ctx, r.querier(ctx), tokenHash)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repositories) GetProfile(ctx context.Context, id uuid.UUID) (FetchProfileResponse, error) {
	res, err := r.user.FetchProfile(ctx, r.querier(ctx), id)
	if err != nil {
		return FetchProfileResponse{}, err
	}
	return res, nil
}
