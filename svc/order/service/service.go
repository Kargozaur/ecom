package service

import (
	"context"
	"errors"
	"order/repo"
	dbresp "order/repo/db_resp"
	"order/types"
	orderv1 "proto/out/order/v1"
	"uuid"
)

type Service struct {
	repo *repo.Repo
}

func NewService(repo *repo.Repo) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetOrder(ctx context.Context, userID, orderID string) (*orderv1.FetchOrderResponse, error) {
	user, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	order, err := uuid.Parse(orderID)
	if err != nil {
		return nil, err
	}
	res, err := s.repo.FetchOrder(ctx, user, order)
	if err != nil {
		return nil, err
	}
	response := &orderv1.FetchOrderResponse{}
	l := len(res.Items)
	response.Items = make([]*orderv1.OrderItem, l)
	for i := range l {
		response.Items[i] = &orderv1.OrderItem{
			Name:     res.Items[i].Name,
			Quantity: int32(res.Items[i].Quantity),
			Price:    float32(res.Items[i].Price),
		}
	}
	response.Status = res.Status
	response.CreatedAt = res.CreatedAt.String()
	return response, nil
}

func (s *Service) GetOrders(ctx context.Context, userID string, page int32) ([]*orderv1.FetchOrdersResponse, error) {
	user, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	limit := page
	offset := ((limit - 1) * 10)
	res, err := s.repo.FetchOrders(ctx, user, limit, offset)
	if err != nil {
		return nil, err
	}
	l := len(res)
	response := make([]*orderv1.FetchOrdersResponse, l)
	for i := range l {
		response[i] = &orderv1.FetchOrdersResponse{
			OrderId:    res[i].OrderID.String(),
			Status:     res[i].Status,
			TotalPrice: float32(res[i].TotalPrice),
		}
	}
	return response, nil
}

func (s *Service) CreateOrder(ctx context.Context, params *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	var txRes *dbresp.CreateOrderResponse
	userID, ok := ctx.Value(types.UserIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("failed to get user id")
	}
	totalPrice := s.calculateTotalPrice(params.GetItems())
	items := s.buildItems(params)
	err := s.repo.WithinTx(ctx, func(c context.Context) error {
		v, err := s.repo.CreateOrder(c, userID, totalPrice, items)
		if err != nil {
			return err
		}
		txRes = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	if txRes == nil {
		return nil, errors.New("order creation returned no result")
	}
	response := &orderv1.CreateOrderResponse{
		OrderId: txRes.ID,
		Status:  txRes.Status,
		Message: "order created successfully",
	}
	return response, nil
}

func (s *Service) calculateTotalPrice(params []*orderv1.OrderItem) float64 {
	var totalPrice float32
	for _, item := range params {
		totalPrice += item.GetPrice() * float32(item.GetQuantity())
	}
	return float64(totalPrice)
}

func (s *Service) buildItems(params *orderv1.CreateOrderRequest) []dbresp.OrderItems {
	res := make([]dbresp.OrderItems, len(params.GetItems()))
	for _, item := range params.GetItems() {
		itemID, err := uuid.Parse(item.GetItemId())
		if err != nil {
			return nil
		}
		res = append(res, dbresp.OrderItems{
			ItemID:    itemID,
			ItemName:  item.GetName(),
			Quantity:  item.GetQuantity(),
			ItemPrice: float64(item.GetPrice()),
		})
	}
	return res
}
