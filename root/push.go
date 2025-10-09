package root

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PKr-Parivar/PKr-Base/config"
	"github.com/PKr-Parivar/PKr-Base/dialer"
	"github.com/PKr-Parivar/PKr-Base/encrypt"
	"github.com/PKr-Parivar/PKr-Base/filetracker"
	"github.com/PKr-Parivar/PKr-Base/pb"
)

func Push(workspace_name, push_desc string) {
	// Getting Workspace's Absolute Path
	workspace_path, err := config.GetSendWorkspaceFilePath(workspace_name)
	if err != nil {
		fmt.Println("Error while getting Absolute Workspace Path:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Creating New Tree
	fmt.Println("Creating File Tree Structure of Workspace ...")
	new_tree, err := config.GetNewTree(workspace_path)
	if err != nil {
		fmt.Println("Could Not Create New Tree of the Current Workspace:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Getting Old Tree
	fmt.Println("Fetching Old Push's File Tree ...")
	old_tree, err := config.ReadFromTreeFile(workspace_path)
	if err != nil {
		fmt.Println("Could Not Read Old Tree of the file_tree.json:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Comparing Old & New Trees
	fmt.Println("Checking for Changes in Workspace ...")
	changes := config.CompareTrees(old_tree, new_tree)
	if len(changes) == 0 {
		fmt.Println("No New Changes Detected in 'PUSH'")
		return
	}
	fmt.Println("Changes're Detected ...")

	// Reading Last Hash from Config
	workspace_conf, err := config.ReadFromWorkspaceConfigFile(filepath.Join(workspace_path, ".PKr", "workspace-config.json"))
	if err != nil {
		fmt.Println("Error while Reading from PKr Config File:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Zip Entire Workspace
	fmt.Println("Zipping Entire Workspace ...")
	zip_destination_path := filepath.Join(workspace_path, ".PKr", "Files", "Current") + string(filepath.Separator)
	err = filetracker.ZipData(workspace_path, zip_destination_path, strconv.Itoa(workspace_conf.LastPushNum+1))
	if err != nil {
		fmt.Println("Error while Creating Zip File of Entire Workspace:", err)
		fmt.Println("Source: Push()")
		return
	}
	fmt.Println("Zip File Created")

	// Generating AES Key
	fmt.Println("Generating & Storing Workspace Keys ...")
	key, err := encrypt.AESGenerakeKey(16)
	if err != nil {
		fmt.Println("Failed to Generate AES Keys:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Storing AES Key
	err = os.WriteFile(zip_destination_path+"AES_KEY", key, 0644)
	if err != nil {
		fmt.Println("Failed to Write AES Key to File:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Generating AES IV
	iv, err := encrypt.AESGenerateIV()
	if err != nil {
		fmt.Println("Failed to Generate IV Keys:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Storing AES IV
	err = os.WriteFile(zip_destination_path+"AES_IV", iv, 0644)
	if err != nil {
		fmt.Println("Failed to Write AES IV to File:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Removing Previous Commit's Entire Zip File
	old_zipped_filepath := zip_destination_path + strconv.Itoa(workspace_conf.LastPushNum) + ".zip"
	err = os.Remove(old_zipped_filepath)
	if err != nil {
		fmt.Println("Error while Deleting Old Zip File:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Writing New Tree to Config
	err = config.WriteToFileTree(workspace_path, new_tree)
	if err != nil {
		fmt.Println("Error Write Tree to file:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Adding Changes to PKr Config
	fmt.Println("Updating New Changes to the Workspace Config ...")
	updates := config.Updates{
		PushNum:  workspace_conf.LastPushNum + 1,
		PushDesc: push_desc,
		Changes:  changes,
	}
	err = config.AppendWorkspaceUpdates(updates, workspace_path)
	if err != nil {
		fmt.Println("Error Write Tree to file:", err)
		fmt.Println("Source: Push()")
		return
	}

	changes_push_range := strconv.Itoa(workspace_conf.LastPushNum) + "-" + strconv.Itoa(workspace_conf.LastPushNum+1)
	changes_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes", changes_push_range)

	src_zip_path := filepath.Join(workspace_path, ".PKr", "Files", "Current", strconv.Itoa(workspace_conf.LastPushNum+1)+".zip")
	dst_zip_path := filepath.Join(workspace_path, ".PKr", "Files", "Changes", changes_push_range, changes_push_range+".zip")

	err = filetracker.CleanFilesFromWorkspace(filepath.Join(workspace_path, ".PKr", "Files", "Changes"))
	if err != nil {
		fmt.Println("Error while Clearing .PKr/Files/Changes:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Create Zip File of Changes
	fmt.Println("Zipping Changes ...")
	err = filetracker.ZipUpdates(changes, src_zip_path, dst_zip_path)
	if err != nil {
		fmt.Println("Error while Creating Zip for Changes:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Create AES Key for Changes Zip
	fmt.Println("Generating & Storing 'Changes' Keys ...")
	changes_key, err := encrypt.AESGenerakeKey(16)
	if err != nil {
		fmt.Println("Failed to Generate AES Keys:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Storing Key
	err = os.WriteFile(filepath.Join(changes_path, "AES_KEY"), changes_key, 0644)
	if err != nil {
		fmt.Println("Failed to Write AES Key to File:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Creating AES IV for Changes Zip
	changes_iv, err := encrypt.AESGenerateIV()
	if err != nil {
		fmt.Println("Failed to Generate IV Keys:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Storing IV
	err = os.WriteFile(filepath.Join(changes_path, "AES_IV"), changes_iv, 0644)
	if err != nil {
		fmt.Println("Failed to Write AES IV to File:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Encrypting Changes Zip
	fmt.Println("Encrypting 'Changes' Zip ...")
	changes_enc_zip_filepath := strings.Replace(dst_zip_path, ".zip", ".enc", 1)
	err = encrypt.EncryptZipFileAndStore(dst_zip_path, changes_enc_zip_filepath, changes_key, changes_iv)
	if err != nil {
		fmt.Println("Error while Encrypting 'Changes' Zip File, Storing it & Deleting Zip File:", err)
		fmt.Println("Source: Push()")
		return
	}

	// Get Details from Config
	// Get Details from user-config
	user_conf, err := config.ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while Reading user-config:", err)
		fmt.Println("Source: Push()")
		return
	}

	// New GRPC Client
	fmt.Println("Registering Push & Notifying Listeners ...")
	gRPC_cli_service_client, err := dialer.GetNewGRPCClient(user_conf.ServerIP)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Create New GRPC Client")
		fmt.Println("Source: Push()")
		return
	}

	// Prepare req
	req := &pb.NotifyNewPushToListenersRequest{
		WorkspaceOwnerUsername: user_conf.Username,
		WorkspaceOwnerPassword: user_conf.Password,
		WorkspaceName:          workspace_name,
		NewWorkspacePushNum:    int32(updates.PushNum),
	}

	// Request Timeout
	ctx, cancelFunc := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancelFunc()

	_, err = gRPC_cli_service_client.NotifyNewPushToListeners(ctx, req)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Notify New Push to Listeners")
		fmt.Println("Source: Push()")
		return
	}

	err = config.UpdateLastPushNum(workspace_name, updates.PushNum)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Add New Push to Config")
		fmt.Println("Source: Push()")
		return
	}
	fmt.Println("Registered New Push Successfully")
}
