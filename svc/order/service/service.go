package service

import (
	"context"
	"errors"
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
	response := s.buildResponseItem(res)
	return response, nil
}

func (s *Service) GetOrders(ctx context.Context, userID string, page int32) ([]*orderv1.FetchOrdersResponse, error) {
	user, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	limit := int32(10)
	offset := (page - 1) * limit
	res, err := s.repo.FetchOrders(ctx, user, limit, offset)
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

func (s *Service) buildResponseItems(items []dbresp.Orders) []*orderv1.FetchOrdersResponse {
	res := make([]*orderv1.FetchOrdersResponse, 0, len(items))
	for _, item := range items {
		res = append(res, &orderv1.FetchOrdersResponse{
			OrderId:    item.OrderID.String(),
			Status:     item.Status,
			TotalPrice: float32(item.TotalPrice),
		})
	}
	return res
}

func (s *Service) buildResponseItem(queryRes *dbresp.FetchOrder) *orderv1.FetchOrderResponse {
	response := &orderv1.FetchOrderResponse{}
	for _, item := range queryRes.Items {
		response.Items = append(response.Items, &orderv1.OrderItem{
			Name:     item.Name,
			Quantity: int32(item.Quantity),
			Price:    float32(item.Price),
		})
	}
	response.Status = queryRes.Status
	response.CreatedAt = queryRes.CreatedAt.String()
	return response
}
