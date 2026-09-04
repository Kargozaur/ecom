package main

import (
	"context"
	"log"
	"order/server"
	"os"
	"os/signal"
	"pkg/envreader"
	orderv1 "proto/out/order/v1"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	url := envreader.Read("ORDER_DB", "postgres://postgres:1234@localhost:5433/order_db?sslmode=disable&pool_max_conn=10")
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
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app, err := NewApp(ctx)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer app.Close()
	if err := app.Run(ctx); err != nil {
		log.Println(err.Error())
	}
}
