package repo

import (
	"context"
	"items/db"
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
)

type items struct{}

func (i *items) GetItem(ctx context.Context, queries *db.Queries, id uuid.UUID) (*db.GetItemRow, error) {
	row, err := queries.GetItem(ctx, pgtype.UUID{
		Bytes: id, Valid: true,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (i *items) GetItems(ctx context.Context, queries *db.Queries, categories []string) ([]db.GetItemsRow, error) {
	rows, err := queries.GetItems(ctx, categories)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
