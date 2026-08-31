package repo

import (
	"context"
	"order/db"
	dbresp "order/repo/db_resp"
	"uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type Repo struct {
	pool           *pgxpool.Pool
	queries        *db.Queries
	orderRepo      *orderRepo
	orderItemsRepo *orderItemsRepo
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{
		queries:        db.New(pool),
		orderRepo:      newOrderRepo(),
		orderItemsRepo: newOrderItemsRepo(),
	}
}

func (r *Repo) querier(ctx context.Context) *db.Queries {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *Repo) WithinTx(ctx context.Context, fn func(context.Context) error) error {
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

func (r *Repo) FetchOrders(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]dbresp.Orders, error) {
	result, err := r.orderRepo.fetchOrders(ctx, r.querier(ctx), userID, limit, offset)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repo) FetchOrder(ctx context.Context, userID, orderID uuid.UUID) (*dbresp.FetchOrder, error) {
	result, err := r.orderRepo.fetchOrder(ctx, r.querier(ctx), userID, orderID)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repo) CreateOrder(ctx context.Context, userID uuid.UUID,
	price float64, itemIDs []uuid.UUID) (*dbresp.CreateOrderResponse, error) {
	row, err := r.orderRepo.createOrder(ctx, r.querier(ctx), userID, price)
	if err != nil {
		return nil, err
	}
	id, err := uuid.Parse(row.ID)
	if err != nil {
		return nil, err
	}
	err = r.orderItemsRepo.insertOrderItems(ctx, r.querier(ctx), id, itemIDs)
	if err != nil {
		return nil, err
	}
	return &dbresp.CreateOrderResponse{
		ID:         id.String(),
		TotalPrice: price,
		Status:     row.Status,
	}, nil
}

func (r *Repo) CancelOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	err := r.orderRepo.cancelOrder(ctx, r.querier(ctx), userID, orderID)
	if err != nil {
		return err
	}
	return nil
}
