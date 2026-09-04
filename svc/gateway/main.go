package main

import (
	"context"
	"gateway/handlers"
	"gateway/handlers/health"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pkg/token"
	userv1 "proto/out/user/v1"
	"syscall"
	"time"
)

func srvConfig(clients *Conns) *http.Server {
	cfg, _ := token.NewTokenConfig()
	validator := token.NewTokenValidator(cfg)
	mux := http.NewServeMux()
	pwdPolicies := CreatePasswordPolicies()
	handlers.RegisterUserHandler(mux, userv1.NewUserServiceClient(clients.userConn), validator, pwdPolicies)
	health.Health(mux)
	timeoutHandler := http.TimeoutHandler(mux, 10*time.Second, "Request timed out")
	srv := &http.Server{
		Addr:              ":8080",
		Handler:           timeoutHandler,
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	return srv
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
		log.Fatalf("Failed to start the server: %s", err.Error())
	}
}
