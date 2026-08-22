package repo

import (
	"context"
	"reg/db"
	"time"
	"uuid"

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

type userRepo struct{}

func (u *userRepo) RegisterUser(ctx context.Context, queries *db.Queries,
	params db.CreateUserParams) (CreateResponse, error) {
	result, err := queries.CreateUser(ctx, params)
	if err != nil {
		return CreateResponse{}, err
	}
	return CreateResponse{
		ID:        result.ID.String(),
		Name:      result.Name,
		CreatedAt: result.CreatedAt.Time,
	}, nil
}

func (u *userRepo) LoginUser(ctx context.Context, queries *db.Queries, email string) (db.FetchPasswordRow, error) {
	res, err := queries.FetchPassword(ctx, email)
	if err != nil {
		return db.FetchPasswordRow{}, err
	}
	return res, nil
}

func (u *userRepo) FetchProfile(ctx context.Context, queries *db.Queries,
	id uuid.UUID) (FetchProfileResponse, error) {
	result, err := queries.FetchProfile(ctx, pgtype.UUID{Bytes: id, Valid: true})
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
