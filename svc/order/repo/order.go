package repo

import (
	"context"
	"encoding/json"
	"order/db"
	dbresp "order/repo/db_resp"
	"sync"
	"uuid"

	"github.com/jackc/pgx/v5/pgtype"
)

type orderRepo struct{}

type orderItemsRepo struct{}

func (o *orderRepo) fetchOrder(ctx context.Context, queries *db.Queries, userID, orderID uuid.UUID) (*dbresp.FetchOrder, error) {
	result, err := queries.FetchOrder(ctx, db.FetchOrderParams{
		ID:     pgtype.UUID{Bytes: orderID, Valid: true},
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	var r dbresp.FetchOrder
	if err := json.Unmarshal(result.Items, &r.Items); err != nil {
		return nil, err
	}
	r.OrderID = orderID
	r.Status = string(result.Status)
	r.TotalPrice = float64(result.TotalPrice.Exp)
	r.CreatedAt = result.CreatedAt.Time
	return &r, nil
}

func (o *orderRepo) fetchOrders(ctx context.Context, queries *db.Queries, userID uuid.UUID, limit, offset int32) ([]dbresp.Orders, error) {
	result, err := queries.FetchUserOrders(ctx, db.FetchUserOrdersParams{
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}
	orders := make([]dbresp.Orders, 0, len(result))
	for _, row := range result {
		id, err := uuid.Parse(row.ID.String())
		if err != nil {
			return nil, err
		}

		price, err := row.TotalPrice.Float64Value()
		if err != nil {
			return nil, err
		}
		orders = append(orders, dbresp.Orders{
			OrderID:    id,
			TotalPrice: price.Float64,
			Status:     string(row.Status),
			CreatedAt:  row.CreatedAt.Time,
		})
	}
	return orders, nil
}

func (o *orderRepo) cancelOrder(ctx context.Context, queries *db.Queries, userID, orderID uuid.UUID) error {
	row, err := queries.SelectOrderForUpdate(ctx, db.SelectOrderForUpdateParams{
		ID:     pgtype.UUID{Bytes: orderID, Valid: true},
		UserID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil || !row.ID.Valid {
		return err
	}
	err = queries.CancelOrder(ctx, db.CancelOrderParams{
		ID:     row.ID,
		UserID: row.UserID,
	})
	if err != nil {
		return err
	}
	return nil
}

func (o *orderRepo) createOrder(ctx context.Context, queries *db.Queries, userID uuid.UUID, totalPrice float64) (dbresp.CreateOrderResponse, error) {
	var x pgtype.Numeric
	if err := x.Scan(totalPrice); err != nil {
		return dbresp.CreateOrderResponse{}, err
	}
	row, err := queries.CreateOrder(ctx, db.CreateOrderParams{
		UserID:     pgtype.UUID{Bytes: userID, Valid: true},
		TotalPrice: x,
	})
	if err != nil {
		return dbresp.CreateOrderResponse{}, err
	}
	resp := dbresp.CreateOrderResponse{
		ID:         row.ID.String(),
		TotalPrice: totalPrice,
		Status:     string(row.Status),
	}
	return resp, nil
}

func (o *orderItemsRepo) insertOrderItems(ctx context.Context, queries *db.Queries, orderID uuid.UUID, itemID []uuid.UUID) error {
	var wg sync.WaitGroup
	l := len(itemID)
	errChan := make(chan error, l)
	for _, id := range itemID {
		wg.Add(1)
		go func(id uuid.UUID) {
			defer wg.Done()
			if err := queries.CreateOrderItems(ctx, db.CreateOrderItemsParams{
				ItemID:  pgtype.UUID{Bytes: id, Valid: true},
				OrderID: pgtype.UUID{Bytes: orderID, Valid: true},
			}); err != nil {
				errChan <- err
			}
		}(id)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		if err != nil {
			return err
		}
	}
	return nil
}
