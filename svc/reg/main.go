package main

import (
	"log"
	"net"
	userv1 "proto/out/user/v1"
	"reg/server"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50000")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer listener.Close()
	grpcServer := grpc.NewServer()
	serv := &server.GRPCServer{}
	userv1.RegisterUserServiceServer(grpcServer, serv)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve %s\n", err.Error())
	}
}
