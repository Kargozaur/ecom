package main

import (
	"context"
	"log"
	"net"
	processor "order/event_processor"
	"order/server"
	"os"
	"os/signal"
	"pkg/envreader"
	orderv1 "proto/out/order/v1"
	"sync"
	"syscall"
	"time"

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
	proc := processor.NewProcessor()
	defer proc.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		proc.Run(ctx)
	}()
	go func() {
		defer wg.Done()
		if err := grpcServer.Serve(listener); err != nil {
			log.Println(err.Error())
		}
	}()
	<-ctx.Done()
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(time.Second * 10):
		log.Println("enforce stop")
		grpcServer.Stop()
	}
	wg.Wait()
}
