package root

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PKr-Parivar/PKr-Base/config"
	"github.com/PKr-Parivar/PKr-Base/dialer"
	"github.com/PKr-Parivar/PKr-Base/encrypt"
	"github.com/PKr-Parivar/PKr-Base/filetracker"
	"github.com/PKr-Parivar/PKr-Base/pb"
)

func InitWorkspace(workspace_password, push_desc string) {
	// Get Curr Directory as Workspace Path
	workspace_path, err := os.Getwd()
	if err != nil {
		fmt.Println("Error Cannot Call Getwd():", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	workspace_config_file_path := filepath.Join(workspace_path, ".PKr", "workspace-config.json")
	_, err = os.Stat(workspace_config_file_path)
	if err == nil {
		fmt.Println("It Seems Workspace is Already Initialized ...")
		return
	} else if os.IsNotExist(err) {
		fmt.Println("Initializing New Workspace ...")
	} else {
		fmt.Println("Error while checking Existence of workspace-config file:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	workspace_path_split := strings.Split(workspace_path, string(filepath.Separator))
	workspace_name := workspace_path_split[len(workspace_path_split)-1]

	// Create the workspace config file
	if err := config.CreatePKRConfigIfNotExits(workspace_name, workspace_path); err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Create .Pkr/PKRConfig.json")
		fmt.Println("Source: InitWorkspace()")
		return
	}

	zip_destination_path := filepath.Join(workspace_path, ".PKr", "Files", "Current") + string(filepath.Separator)

	// Generating Key
	fmt.Println("Generating & Storing Workspace Keys ...")
	key, err := encrypt.AESGenerakeKey(16)
	if err != nil {
		fmt.Println("Failed to Generate AES Keys:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Storing Key
	err = os.WriteFile(zip_destination_path+"AES_KEY", key, 0644)
	if err != nil {
		fmt.Println("Failed to Write AES Key to File:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Generating IV
	iv, err := encrypt.AESGenerateIV()
	if err != nil {
		fmt.Println("Failed to Generate IV Keys:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Storing IV
	err = os.WriteFile(zip_destination_path+"AES_IV", iv, 0644)
	if err != nil {
		fmt.Println("Failed to Write AES IV to File:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Compute Tree First
	// Will reduce Syscall times for zipdata
	// OS will cache files in memory -- Thus reducing time taken to zip data
	// Create Tree
	fmt.Println("Creating & Storing Workspace File Structure Tree ...")
	tree, err := config.GetNewTree(workspace_path)
	if err != nil {
		fmt.Println("Error Could not Create Tree:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Store Tree
	err = config.WriteToFileTree(workspace_path, tree)
	if err != nil {
		fmt.Println("Error Write Tree to file:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Write Updates in PKr Config
	changes := config.CompareTrees(config.FileTree{}, tree)
	updates := config.Updates{
		Changes:  changes,
		PushNum:  0,
		PushDesc: push_desc,
	}

	err = config.AppendWorkspaceUpdates(updates, workspace_path)
	if err != nil {
		fmt.Println("Error while Adding Changes to PKr Config:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Creating Zip File of Entire Workspace
	fmt.Println("Creating Zip of Entire Workspace ...")
	err = filetracker.ZipData(workspace_path, zip_destination_path, "0")
	if err != nil {
		fmt.Println("Error while Getting Hash of Zipped Data:", err)
		fmt.Println("Source InitWorkspace()")
		return
	}

	// Register the workspace in the main userConfig file
	if err := config.RegisterNewSendWorkspace(workspace_name, workspace_path, workspace_password); err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Register Workspace to userConfig File")
		fmt.Println("Source: InitWorkspace()")
		return
	}

	err = config.UpdateLastPushNum(workspace_name, 0)
	if err != nil {
		fmt.Println("Error while Updating Last to Config:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Get Details from user-config
	user_conf, err := config.ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while Reading user-config:", err)
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Create New gRPC Client
	fmt.Println("Sending Init Workspace Request to Server ....")
	gRPC_cli_service_client, err := dialer.GetNewGRPCClient(user_conf.ServerIP)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Create New GRPC Client")
		fmt.Println("Source: InitWorkspace()")
		return
	}

	// Prepare gRPC Request
	req := &pb.RegisterWorkspaceRequest{
		Username:      user_conf.Username,
		Password:      user_conf.Password,
		WorkspaceName: workspace_name,
	}

	// Request Timeout
	ctx, cancelFunc := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancelFunc()

	// Sending Request ...
	_, err = gRPC_cli_service_client.RegisterWorkspace(ctx, req)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Register User")
		fmt.Println("Source: InitWorkspace()")
		return
	}

	fmt.Println("Workspace Initialized Successfully")
}
