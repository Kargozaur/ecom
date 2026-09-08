package main

import (
	"context"
	"errors"
	"log"
	"net"
	"pkg/envreader"
	userv1 "proto/out/user/v1"
	"reg/server"
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

func NewApp(ctx context.Context, addr string) (*App, error) {
	pool, err := initDB(ctx)
	if err != nil {
		return nil, errors.New("init db: " + err.Error())
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		pool.Close()
		return nil, errors.New("listen: " + err.Error())
	}

	grpcServer, err := initServer(pool)
	if err != nil {
		listener.Close()
		pool.Close()
		return nil, errors.New("init server: " + err.Error())
	}

	return &App{
		pool:       pool,
		listener:   listener,
		grpcServer: grpcServer,
	}, nil
}

func (a *App) Close() {
	a.listener.Close()
	a.pool.Close()
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
