package server

import orderv1 "proto/out/order/v1"

type GRPCServer struct {
	orderv1.UnimplementedOrderServiceServer
}
