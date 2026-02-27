package root

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PKr-Parivar/PKr-Base/config"
	"github.com/PKr-Parivar/PKr-Base/dialer"
	"github.com/PKr-Parivar/PKr-Base/pb"
	"github.com/PKr-Parivar/PKr-Base/utils"
)

const CONTEXT_TIMEOUT = 60 * time.Second

func Install(server_ip, username, password string) {
	out := strings.Split(server_ip, ":")
	grpcPort, err := strconv.Atoi(out[1])
	if err != nil {
		fmt.Println("Error while converting grpc port addr to int:", err)
		fmt.Println("Source: Install()")
		return
	}

	// This is now serverip only - eg. 192.168.1.1
	server_ip = out[0]

	user_config_file_path, err := utils.GetUserConfigFilePath()
	if err != nil {
		fmt.Println("Error while Getting Path of user-config:", err)
		fmt.Println("Source: Install()")
		return
	}

	_, err = os.Stat(user_config_file_path)
	if err == nil {
		fmt.Println("It Seems PKr is Already Installed...")
		return
	} else if os.IsNotExist(err) {
		fmt.Println("Installing PKr ...")
	} else {
		fmt.Println("Error while checking Existence of user-config file:", err)
		fmt.Println("Source: Install()")
		return
	}

	if server_ip == "" || username == "" || password == "" {
		fmt.Println("Username or Password or Server IP MUST NOT be Empty")
		return
	}

	fmt.Println("Registering User, Sending Request to Server ...")
	// New GRPC Client
	grpcAddr := fmt.Sprintf("%s:%d", server_ip, grpcPort)
	gRPC_cli_service_client, err := dialer.GetNewGRPCClient(grpcAddr)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Create New GRPC Client")
		fmt.Println("Source: Install()")
		return
	}

	// Prepare req
	req := &pb.RegisterRequest{
		Username: username,
		Password: password,
	}

	// Request Timeout
	ctx, cancelFunc := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancelFunc()

	// Sending Request ...
	resp, err := gRPC_cli_service_client.Register(ctx, req)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Register User")
		fmt.Println("Source: Install()")
		return
	}

	if int(resp.WsPort) == 0 {
		fmt.Println("[WARNING] returned ws_port by server is 0. Source: Install()")
	}

	// Add Credentials to Config
	err = config.CreateUserConfigIfNotExists(username, password, server_ip, grpcPort, int(resp.WsPort))
	if err != nil {
		fmt.Println("Error while Adding Credentials to user-config:", err)
		fmt.Println("Source: Install()")
		return
	}
	fmt.Println("PKr Installed Successfully")
}
