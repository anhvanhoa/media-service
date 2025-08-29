package main

import (
	"context"
	"log"
	"media-service/bootstrap"
)

func main() {
	app := bootstrap.NewApp()
	grpcServer := app.Start()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := grpcServer.Start(ctx); err != nil {
		log.Fatal("gRPC server error: " + err.Error())
	}
}
