package repo

import (
	"context"
	"reg/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type refreshRepo struct{}

func (r *refreshRepo) CreateToken(ctx context.Context, queries *db.Queries,
	id uuid.UUID,
	tokenHash string) error {
	params := db.CreateTokenParams{UserID: pgtype.UUID{Bytes: id, Valid: true},
		TokenHash: pgtype.Text{String: tokenHash, Valid: true}}
	err := queries.CreateToken(ctx, params)
	if err != nil {
		return err
	}
	return nil
}

func (r *refreshRepo) DeleteToken(ctx context.Context, queries *db.Queries,
	tokenHash string) error {
	row, err := queries.GetTokenForUpdate(ctx, pgtype.Text{String: tokenHash, Valid: true})
	err = queries.DeleteToken(ctx, db.DeleteTokenParams{TokenHash: row.TokenHash, UserID: row.UserID})
	if err != nil {
		return err
	}
	return nil
}
