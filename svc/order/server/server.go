package server

import (
	"context"
	"errors"
	processor "order/event_processor"
	"order/repo"
	"order/service"
	orderv1 "proto/out/order/v1"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	service *service.Service
	orderv1.UnimplementedOrderServiceServer
}

func NewGRPCServer(pool *pgxpool.Pool, proc *processor.Processor) *GRPCServer {
	return &GRPCServer{
		service: service.NewService(repo.NewRepo(pool), proc),
	}
}

func (s *GRPCServer) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	order, err := s.service.CreateOrder(ctx, req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return order, nil
}

func (s *GRPCServer) FetchOrder(ctx context.Context, req *orderv1.FetchOrderRequest) (*orderv1.FetchOrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	order, err := s.service.GetOrder(ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return order, nil
}

func (s *GRPCServer) FetchOrders(ctx context.Context, req *orderv1.FetchOrdersRequest) (*orderv1.FetchOrdersResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	orders, err := s.service.GetOrders(ctx, req.GetPage())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return orders, nil
}

func (s *GRPCServer) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	order, err := s.service.CancelOrder(ctx, req.OrderId)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return order, nil
}
