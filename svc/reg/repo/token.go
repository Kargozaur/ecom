package repo

import (
	"context"
	"reg/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type refreshRepo struct {
	db      *pgx.Conn
	queries *db.Queries
}

func (r *refreshRepo) CreateToken(ctx context.Context, id uuid.UUID, tokenHash string) error {
	tx, _ := r.db.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)
	params := db.CreateTokenParams{UserID: pgtype.UUID{Bytes: id, Valid: true},
		TokenHash: pgtype.Text{String: tokenHash, Valid: true}}
	err := r.queries.CreateToken(ctx, params)
	if err != nil {
		return err
	}
	tx.Commit(ctx)
	return nil
}

func (r *refreshRepo) DeleteToken(ctx context.Context, tokenHash string) error {
	tx, _ := r.db.BeginTx(ctx, pgx.TxOptions{})
	defer tx.Rollback(ctx)
	row, err := r.queries.GetTokenForUpdate(ctx, pgtype.Text{String: tokenHash, Valid: true})
	err = r.queries.DeleteToken(ctx, db.DeleteTokenParams{TokenHash: row.TokenHash, UserID: row.UserID})
	if err != nil {
		return err
	}
	tx.Commit(ctx)
	return nil
}
