package server

import (
	processor "order/event_processor"
	"order/repo"
	"order/service"
	orderv1 "proto/out/order/v1"

	"github.com/jackc/pgx/v5/pgxpool"
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
