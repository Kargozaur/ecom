package main

import (
	userv1 "proto/out/user/v1"

	"google.golang.org/grpc"
)

type grpcClients struct {
	userClient userv1.UserServiceClient
}

// User client has index 0
func initClients(conns ...*grpc.ClientConn) *grpcClients {
	return &grpcClients{
		userClient: userv1.NewUserServiceClient(conns[0]),
	}
}
