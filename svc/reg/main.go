package main

import (
	"context"
	"log"
	"net"
	"pkg/envreader"
	userv1 "proto/out/user/v1"
	"reg/server"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	url := envreader.Read("DB_URL", "postgres://postgres:1234@localhost:5433/user_db?sslmode=disable&pool_max_conn=10")
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func main() {
	ctx := context.Background()
	pool, err := initDB(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer pool.Close()
	listener, err := net.Listen("tcp", ":50000")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer listener.Close()
	grpcServer := grpc.NewServer()
	serv := server.NewGRPCServer(pool)
	if serv == nil {
		log.Fatal("Failed to create GRPC server")
	}
	userv1.RegisterUserServiceServer(grpcServer, serv)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve %s\n", err.Error())
	}
}
