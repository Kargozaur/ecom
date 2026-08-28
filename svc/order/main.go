package main

import (
	"context"
	"log"
	"net"
	"order/server"
	"pkg/envreader"
	orderv1 "proto/out/order/v1"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	url := envreader.Read("ORDER_URL", "postgres://postgres:1234@localhost:5433/order_db?sslmode=disable&pool_max_conn=10")
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func initServer() *grpc.Server {
	grpcServer := grpc.NewServer()
	srv := &server.GRPCServer{}
	orderv1.RegisterOrderServiceServer(grpcServer, srv)
	return grpcServer
}

func main() {
	ctx := context.Background()
	pool, err := initDB(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer pool.Close()
	listener, err := net.Listen("tcp", ":50001")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer listener.Close()
	grpcServer := initServer()
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatal(err.Error())
	}
}
