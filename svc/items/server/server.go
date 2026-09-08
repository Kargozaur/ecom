package server

import itemsv1 "proto/out/items/v1"

type GRPCServer struct {
	itemsv1.UnimplementedItemsServiceServer
}

func NewGRPCServer() *GRPCServer {
	return &GRPCServer{}
}
