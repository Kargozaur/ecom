package main

import (
	"context"
	"items/server"
	"log"
	"net"
	"pkg/envreader"
	itemsv1 "proto/out/items/v1"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type App struct {
	pool       *pgxpool.Pool
	listener   net.Listener
	grpcServer *grpc.Server
}

func initDB(ctx context.Context) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, envreader.Read("ITEMS_DB", "postgres://postgres:1234@localhost:5433/items_db?sslmode=disable&pool_max_conn=10"))
	if err != nil {
		log.Fatal(err.Error())
	}
	return pool
}

func initServer(pool *pgxpool.Pool) *grpc.Server {
	grpcServer := grpc.NewServer()
	itemsv1.RegisterItemsServiceServer(grpcServer, server.NewGRPCServer(pool))
	return grpcServer
}

func NewApp(ctx context.Context) *App {
	pool := initDB(ctx)
	listener, err := net.Listen("tcp", ":5003")
	if err != nil {
		pool.Close()
		log.Fatal(err.Error())
	}
	grpcServer := initServer(pool)
	return &App{pool: pool, listener: listener, grpcServer: grpcServer}
}

func (a *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return a.grpcServer.Serve(a.listener)
	})
	g.Go(func() error {
		<-gctx.Done()
		stopped := make(chan struct{})
		go func() {
			a.grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-time.After(time.Second * 10):
			log.Println("enforced server stop")
			a.grpcServer.Stop()
		}
		return nil
	})
	return g.Wait()
}

func (a *App) Close() error {
	a.pool.Close()
	listenerErr := a.listener.Close()
	return listenerErr
}
