package server

import (
	"context"
	"errors"
	"items/service"
	itemsv1 "proto/out/items/v1"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GRPCServer struct {
	itemsv1.UnimplementedItemsServiceServer
	service service.Service
}

func NewGRPCServer(pool *pgxpool.Pool) *GRPCServer {
	return &GRPCServer{service: service.NewService(pool)}
}

func (s *GRPCServer) GetItem(ctx context.Context, req *itemsv1.GetItemRequest) (*itemsv1.GetItemResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	item, err := s.service.GetItemById(ctx, req.ItemId)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return item, nil
}

func (s *GRPCServer) GetItems(ctx context.Context, req *itemsv1.GetItemsRequest) (*itemsv1.GetItemsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	items, err := s.service.GetItems(ctx, req.Categories)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return items, nil
}
