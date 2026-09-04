package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"pkg/envreader"
	userv1 "proto/out/user/v1"
	"reg/server"
	"syscall"

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

func initServer(pool *pgxpool.Pool) (*grpc.Server, error) {
	srv, err := server.NewGRPCServer(pool)
	if err != nil {
		return nil, err
	}
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, srv)
	return grpcServer, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app, err := NewApp(ctx, ":50000")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer app.Close()
	if err := app.Run(ctx); err != nil {
		log.Println(err.Error())
	}
}
