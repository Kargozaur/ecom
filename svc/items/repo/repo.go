package repo

import (
	"context"
	"items/db"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	items   items
	queries *db.Queries
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		items:   items{},
		queries: db.New(pool),
	}
}

func (r *Repo) GetItems(ctx context.Context, categories []string) ([]db.GetItemsRow, error) {
	items, err := r.items.GetItems(ctx, r.queries, categories)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repo) GetItemById(ctx context.Context, id uuid.UUID) (*db.GetItemRow, error) {
	items, err := r.items.GetItem(ctx, r.queries, id)
	if err != nil {
		return nil, err
	}
	return items, nil
}
