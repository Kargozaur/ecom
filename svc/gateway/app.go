package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

type App struct {
	clients *Conns
	srv     *http.Server
}

func NewApp(ctx context.Context) (*App, error) {
	clients := initClients(ctx)
	if clients == nil {
		return nil, errors.New("failed to initialize grpc clients")
	}
	srv := srvConfig(clients)
	return &App{
		clients: clients,
		srv:     srv,
	}, nil
}

func (a *App) Close() {
	a.clients.Close()
}

func (a *App) Run(ctx context.Context) error {
	serveErrCh := make(chan error, 1)
	go func() {
		if err := a.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErrCh <- err
			return
		}
		serveErrCh <- nil
	}()
	select {
	case err := <-serveErrCh:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.srv.Shutdown(shutdownCtx); err != nil {
		log.Println("server shutdown failed: ", err.Error())
		if closeErr := a.srv.Close(); closeErr != nil {
			return errors.New("failed to close server: " + closeErr.Error())
		}
	}

	return <-serveErrCh
}
