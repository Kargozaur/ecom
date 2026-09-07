package service

import (
	"context"
	"errors"
	"order/db"
	processor "order/event_processor"
	"order/repo"
	dbresp "order/repo/db_resp"
	"order/types"
	orderv1 "proto/out/order/v1"
	"uuid"

	"google.golang.org/protobuf/proto"
)

type Service struct {
	repo *repo.Repo
	proc *processor.Processor
}

func NewService(repo *repo.Repo, proc *processor.Processor) *Service {
	return &Service{repo: repo, proc: proc}
}

func (s *Service) GetOrder(ctx context.Context, orderID string) (*orderv1.FetchOrderResponse, error) {
	userID, ok := ctx.Value(types.UserIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("failed to get user id")
	}
	order, err := uuid.Parse(orderID)
	if err != nil {
		return nil, err
	}
	res, err := s.repo.FetchOrder(ctx, userID, order)
	if err != nil {
		return nil, err
	}
	response := s.buildResponseItem(res)
	return response, nil
}

func (s *Service) GetOrders(ctx context.Context, page int32) (*orderv1.FetchOrdersResponse, error) {
	userID, ok := ctx.Value(types.UserIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("failed to get user id")
	}
	limit := int32(10)
	offset := (page - 1) * limit
	res, err := s.repo.FetchOrders(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	response := s.buildResponseItems(res)
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
		id, err := uuid.Parse(v.ID)
		if err != nil {
			return err
		}
		err = s.repo.CreateEvent(c, id)
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
	}
	go func() {
		body, _ := proto.Marshal(response)
		s.proc.Append([]byte(txRes.ID), body)
	}()
	return response, nil
}

func (s *Service) CancelOrder(ctx context.Context, orderID string) (*orderv1.CancelOrderResponse, error) {
	userID, ok := ctx.Value(types.UserIDKey{}).(uuid.UUID)
	if !ok {
		return nil, errors.New("failed to get user id")
	}
	id, err := uuid.Parse(orderID)
	if err != nil {
		return nil, err
	}
	err = s.repo.WithinTx(ctx, func(c context.Context) error {
		return s.repo.CancelOrder(c, userID, id)
	})
	if err != nil {
		return nil, err
	}
	return &orderv1.CancelOrderResponse{
		Status: string(db.OrderStatusCancelled),
	}, nil
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

func (s *Service) buildResponseItems(items []dbresp.Orders) *orderv1.FetchOrdersResponse {
	res := &orderv1.FetchOrdersResponse{
		Orders: make([]*orderv1.Order, 0, len(items)),
	}
	for _, item := range items {
		res.Orders = append(res.Orders, &orderv1.Order{
			OrderId:    item.OrderID.String(),
			Status:     item.Status,
			TotalPrice: float32(item.TotalPrice),
			CreatedAt:  item.CreatedAt.String(),
		})
	}
	return res
}

func (s *Service) buildResponseItem(queryRes *dbresp.FetchOrder) *orderv1.FetchOrderResponse {
	response := &orderv1.FetchOrderResponse{
		Items:      make([]*orderv1.OrderItem, 0, len(queryRes.Items)),
		TotalPrice: float32(queryRes.TotalPrice),
		Status:     queryRes.Status,
		CreatedAt:  queryRes.CreatedAt.String(),
	}
	for _, item := range queryRes.Items {
		response.Items = append(response.Items, &orderv1.OrderItem{
			Name:     item.Name,
			Quantity: int32(item.Quantity),
			Price:    float32(item.Price),
		})
	}
	return response
}
