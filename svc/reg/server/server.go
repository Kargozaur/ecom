package server

import userv1 "proto/out/user/v1"

type GRPCServer struct {
	userv1.UnimplementedUserServiceServer
}
