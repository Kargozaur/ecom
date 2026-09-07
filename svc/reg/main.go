package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	app, err := NewApp(ctx, ":50000")
	if err != nil {
		log.Fatal(err.Error())
	}
	defer app.Close()
	if err := app.Run(ctx); err != nil {
		log.Println(err.Error())
	}
}
