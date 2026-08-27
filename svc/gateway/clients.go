package main

import (
	"context"
	"pkg/envreader"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initClients(ctx context.Context) *Conns {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	registry := &Conns{}
	g.Go(func() error {
		conn, err := grpc.NewClient(envreader.Read("USER_CONN", "localhost:50000"),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(CreateGRPCRetryPolicy()))
		if err != nil {
			return err
		}
		registry.userConn = conn
		return nil
	})
	g.Go(func() error {
		conn, err := grpc.NewClient(envreader.Read("ORDER_CONN", "localhost:50001"),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(CreateGRPCRetryPolicy()))
		if err != nil {
			return err
		}
		registry.orderConn = conn
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil
	}
	return registry
}
