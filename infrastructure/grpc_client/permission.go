package grpc_client

import (
	gc "github.com/anhvanhoa/service-core/domain/grpc_client"
	proto_permission "github.com/anhvanhoa/sf-proto/gen/permission/v1"
)

type PermissionClientImpl struct {
	PermissionServiceClient proto_permission.PermissionServiceClient
}

func NewPermissionClient(client *gc.Client) *PermissionClientImpl {
	return &PermissionClientImpl{
		PermissionServiceClient: proto_permission.NewPermissionServiceClient(client.GetConnection()),
	}
}
