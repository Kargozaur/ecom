package main

import (
	"context"
	"errors"
	"log"
	"net"
	processor "order/event_processor"
	"order/interceptor"
	"order/server"
	"pkg/envreader"
	"pkg/token"
	orderv1 "proto/out/order/v1"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type App struct {
	pool       *pgxpool.Pool
	listener   net.Listener
	grpcServer *grpc.Server
	proc       *processor.Processor
}

func initDB(ctx context.Context) (*pgxpool.Pool, error) {
	url := envreader.Read("ORDER_DB", "postgres://postgres:1234@localhost:5433/order_db?sslmode=disable&pool_max_conn=10")
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

func initServer(pool *pgxpool.Pool, proc *processor.Processor) *grpc.Server {
	validator, err := newTokenValidator()
	if err != nil {
		log.Fatal(err.Error())
	}
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptor.TokenInterceptor(validator)))
	srv := server.NewGRPCServer(pool, proc)
	orderv1.RegisterOrderServiceServer(grpcServer, srv)
	return grpcServer
}

func newTokenValidator() (*token.TokenValidator, error) {
	config, err := token.NewTokenConfig()
	if err != nil {
		return nil, err
	}
	return token.NewTokenValidator(config), nil
}

func NewApp(ctx context.Context) (*App, error) {
	pool, err := initDB(ctx)
	if err != nil {
		return nil, errors.New("init db: " + err.Error())
	}

	listener, err := net.Listen("tcp", ":50001")
	if err != nil {
		pool.Close()
		return nil, errors.New("listener: " + err.Error())
	}
	proc := processor.NewProcessor()
	grpcServer := initServer(pool, proc)

	return &App{
		pool:       pool,
		listener:   listener,
		grpcServer: grpcServer,
		proc:       proc,
	}, nil
}

func (a *App) Close() error {
	procErr := a.proc.Close()
	listenerErr := a.listener.Close()
	a.pool.Close()
	return errors.Join(procErr, listenerErr)
}

func (a *App) Run(ctx context.Context) error {
	var wg sync.WaitGroup
	wg.Add(2)
	var grpcErr error
	go func() {
		defer wg.Done()
		a.proc.Run(ctx)
	}()
	go func() {
		defer wg.Done()
		if err := a.grpcServer.Serve(a.listener); err != nil {
			grpcErr = err
		}
	}()
	<-ctx.Done()
	stopped := make(chan struct{})
	go func() {
		a.grpcServer.GracefulStop()
		close(stopped)
	}()
	select {
	case <-stopped:
	case <-time.After(10 * time.Second):
		log.Println("enforced server shutdown")
		a.grpcServer.Stop()
	}

	wg.Wait()
	return grpcErr
}
