package repo

import (
	"context"
	"reg/db"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateResponse struct {
	ID        string
	Name      string
	CreatedAt time.Time
}

type TokenResponse struct {
	Access  string
	Refresh string
	Type    string
}

type FetchProfileResponse struct {
	ID        string
	Name      string
	Email     string
	CreatedAt time.Time
}

type userRepo struct {
	db      *pgx.Conn
	queries *db.Queries
}

func (u *userRepo) RegisterUser(ctx context.Context, params db.CreateUserParams) (CreateResponse, error) {
	tx, err := u.db.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)
	result, err := u.queries.CreateUser(ctx, params)
	if err != nil {
		return CreateResponse{}, err
	}
	tx.Commit(ctx)
	return CreateResponse{
		ID:        result.ID.String(),
		Name:      result.Name,
		CreatedAt: result.CreatedAt.Time,
	}, nil
}

func (u *userRepo) LoginUser(ctx context.Context, email string) (db.FetchPasswordRow, error) {
	res, err := u.queries.FetchPassword(ctx, email)
	if err != nil {
		return db.FetchPasswordRow{}, err
	}
	return res, nil
}

func (u *userRepo) FetchProfile(ctx context.Context, id uuid.UUID) (FetchProfileResponse, error) {
	result, err := u.queries.FetchProfile(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return FetchProfileResponse{}, err
	}
	return FetchProfileResponse{
		ID:        result.ID.String(),
		Name:      result.Name,
		Email:     result.Email,
		CreatedAt: result.CreatedAt.Time,
	}, nil
}
