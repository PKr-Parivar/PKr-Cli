package root

import (
	"context"
	"fmt"

	"github.com/PKr-Parivar/PKr-Base/config"
	"github.com/PKr-Parivar/PKr-Base/dialer"
	"github.com/PKr-Parivar/PKr-Base/pb"
)

func ListAllWorkspaces() {
	// Get Details from user-config
	user_conf, err := config.ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while Reading user-config:", err)
		fmt.Println("Source: ListAllWorkspaces()")
		return
	}

	// New GRPC Client
	gRPC_cli_service_client, err := dialer.GetNewGRPCClient(user_conf.ServerIP)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Create New GRPC Client")
		fmt.Println("Source: ListAllWorkspaces()")
		return
	}

	// Prepare req
	req := &pb.GetAllWorkspacesRequest{
		Username: user_conf.Username,
		Password: user_conf.Password,
	}

	// Request Timeout
	ctx, cancelFunc := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancelFunc()

	res, err := gRPC_cli_service_client.GetAllWorkspaces(ctx, req)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Notify New Push to Listeners")
		fmt.Println("Source: ListAllWorkspaces()")
		return
	}

	for _, workspace := range res.Workspaces {
		fmt.Printf("Workspace Owner: %s, Workspace Name: %s\n", workspace.WorkspaceOwner, workspace.WorkspaceName)
	}
	fmt.Println("Workspace Fetched Successfully")
}
