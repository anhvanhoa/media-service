package main

import (
	"context"
	"log"
	"media-service/bootstrap"
	"media-service/infrastructure/grpc_client"

	gc "github.com/anhvanhoa/service-core/domain/grpc_client"
)

func main() {
	app := bootstrap.NewApp()
	env := app.Env
	clientFactory := gc.NewClientFactory(env.GrpcClients...)
	permissionClient := grpc_client.NewPermissionClient(clientFactory.GetClient(env.PermissionServiceAddr))
	grpcServer := app.Start()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	permissions := app.Helper.ConvertResourcesToPermissions(grpcServer.GetResources())
	if _, err := permissionClient.PermissionServiceClient.RegisterPermission(ctx, permissions); err != nil {
		log.Fatal("Failed to register permission: " + err.Error())
	}
	if err := grpcServer.Start(ctx); err != nil {
		log.Fatal("gRPC server error: " + err.Error())
	}
}
