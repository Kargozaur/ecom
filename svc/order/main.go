package main

import (
	"log"
	"net"
	"order/server"
	orderv1 "proto/out/order/v1"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", ":50001")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer listener.Close()
	srv := &server.GRPCServer{}
	grpcServer := grpc.NewServer()

	orderv1.RegisterOrderServiceServer(grpcServer, srv)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err.Error())
	}
}
