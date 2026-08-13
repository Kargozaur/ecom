package main

import (
	"gateway/handlers"
	"gateway/handlers/health"
	"log"
	"net/http"
	"pkg/token"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func srvConfig(clients *grpcClients) *http.Server {
	cfg, _ := token.NewTokenConfig()
	validator := token.NewTokenValidator(cfg)
	mux := http.NewServeMux()
	handlers.RegisterUserHandler(mux, clients.userClient, validator)
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
	userConn, err := grpc.NewClient("localhost:50000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer userConn.Close()
	clients := initClients(userConn)
	srv := srvConfig(clients)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start the server: %s", err.Error())
	}
}
