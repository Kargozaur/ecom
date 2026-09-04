package main

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
)

type App struct {
	pool       *pgxpool.Pool
	listener   net.Listener
	grpcServer *grpc.Server
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
	var wg sync.WaitGroup
	wg.Add(1)

	var serveErr error
	go func() {
		defer wg.Done()
		if err := a.grpcServer.Serve(a.listener); err != nil {
			serveErr = err
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
	if serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
		return errors.New("serve: " + serveErr.Error())
	}
	return nil
}
