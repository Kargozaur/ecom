package main

import (
	"context"
	"errors"
	"log"
	"net"
	processor "order/event_processor"
	"pkg/token"
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

	grpcServer := initServer()
	proc := processor.NewProcessor()

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
