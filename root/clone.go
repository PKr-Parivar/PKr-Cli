package root

import (
	"bufio"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PKr-Parivar/PKr-Base/config"
	"github.com/PKr-Parivar/PKr-Base/dialer"
	"github.com/PKr-Parivar/PKr-Base/encrypt"
	"github.com/PKr-Parivar/PKr-Base/filetracker"
	"github.com/PKr-Parivar/PKr-Base/handler"
	"github.com/PKr-Parivar/PKr-Base/models"
	"github.com/PKr-Parivar/PKr-Base/pb"
	"github.com/PKr-Parivar/PKr-Base/utils"

	"github.com/PKr-Parivar/kcp-go"
)

const DATA_CHUNK = handler.DATA_CHUNK
const FLUSH_AFTER_EVERY_X_MB = handler.FLUSH_AFTER_EVERY_X_MB

func connectToAnotherUser(workspace_name, workspace_owner_username, server_ip, username, password string) (string, string, *net.UDPConn, *kcp.UDPSession, error) {
	local_port := rand.Intn(16384) + 16384
	fmt.Println("My Local Port:", local_port)

	// Get My Public IP
	my_public_IP, err := dialer.GetMyPublicIP(local_port)
	if err != nil {
		fmt.Println("Error while Getting my Public IP:", err)
		fmt.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err
	}
	fmt.Println("My Public IP Addr:", my_public_IP)

	my_public_IP_split := strings.Split(my_public_IP, ":")
	my_public_IP_only := my_public_IP_split[0]
	my_public_port_only := my_public_IP_split[1]

	my_private_ip, err := dialer.GetMyPrivateIP()
	if err != nil {
		fmt.Println("Error while Getting My Private IP:", err)
		fmt.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err
	}

	// New GRPC Client
	gRPC_cli_service_client, err := dialer.GetNewGRPCClient(server_ip)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Create New GRPC Client")
		fmt.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err
	}

	// Prepare req
	req := &pb.RequestPunchFromReceiverRequest{
		WorkspaceOwnerUsername: workspace_owner_username,
		ListenerUsername:       username,
		ListenerPassword:       password,
		ListenerPublicIp:       my_public_IP_only,
		ListenerPublicPort:     my_public_port_only,
		ListenerPrivateIp:      my_private_ip,
		ListenerPrivatePort:    strconv.Itoa(local_port),
		WorkspaceName:          workspace_name,
	}

	// Request Timeout
	ctx, cancelFunc := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancelFunc()

	// Sending Request to Server
	res, err := gRPC_cli_service_client.RequestPunchFromReceiver(ctx, req)
	if err != nil {
		fmt.Println("Error while Requesting Punch from Receiver:", err)
		fmt.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err

	}
	fmt.Println("Remote Public Addr:", res.WorkspaceOwnerPublicIp+":"+res.WorkspaceOwnerPublicPort)

	// Creating UDP Conn to Perform UDP NAT Hole Punching
	udp_conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: local_port,
		IP:   net.IPv4zero, // or nil
	})
	if err != nil {
		fmt.Printf("Error while Listening to %d: %v\n", local_port, err)
		fmt.Println("Source: connectToAnotherUser()")
		return "", "", nil, nil, err
	}

	var workspace_owner_ip, client_handler_name string
	if res.WorkspaceOwnerPublicIp == my_public_IP_only {
		fmt.Println("Sending Request via Private IP ...")
		workspace_owner_ip = res.WorkspaceOwnerPrivateIp + ":" + res.WorkspaceOwnerPrivatePort
	} else {
		fmt.Println("Sending Request via Public IP ...")
		workspace_owner_ip = res.WorkspaceOwnerPublicIp + ":" + res.WorkspaceOwnerPublicPort
	}

	client_handler_name, err = dialer.WorkspaceListenerUdpNatHolePunching(udp_conn, workspace_owner_ip)
	if err != nil {
		fmt.Println("Error while UDP NAT Hole Punching:", err)
		fmt.Println("Source: connectToAnotherUser()")
		udp_conn.Close()
		return "", "", nil, nil, err
	}
	fmt.Println("UDP NAT Hole Punching Completed Successfully")

	// Creating KCP-Conn, KCP = Reliable UDP
	kcp_conn, err := kcp.DialWithConnAndOptions(workspace_owner_ip, nil, 0, 0, udp_conn)
	if err != nil {
		fmt.Println("Error while Dialing KCP Connection to Remote Addr:", err)
		fmt.Println("Source: connectToAnotherUser()")
		udp_conn.Close()
		return "", "", nil, nil, err
	}

	// KCP Params for Congestion Control
	kcp_conn.SetWindowSize(128, 1024)
	kcp_conn.SetNoDelay(1, 10, 2, 1)
	kcp_conn.SetACKNoDelay(true)
	kcp_conn.SetDSCP(46)

	return client_handler_name, workspace_owner_ip, udp_conn, kcp_conn, nil
}

func fetchAndStoreDataIntoWorkspace(workspace_owner_ip, workspace_name string, udp_conn *net.UDPConn, res models.GetMetaDataResponse) error {
	// Decrypting AES Key
	key, err := encrypt.RSADecryptData(string(res.KeyBytes))
	if err != nil {
		fmt.Println("Error while Decrypting Key:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// Decrypting AES IV
	iv, err := encrypt.RSADecryptData(string(res.IVBytes))
	if err != nil {
		fmt.Println("Error while Decrypting 'IV':", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	workspace_path := "."
	zip_file_path := filepath.Join(workspace_path, ".PKr", "Contents", strconv.Itoa(res.LastPushNum)+".zip")

	// Create Zip File
	zip_file_obj, err := os.Create(zip_file_path)
	if err != nil {
		fmt.Println("Failed to Open & Create Zipped File:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	defer zip_file_obj.Close()

	// To Write Decrypted Data in Chunks
	writer := bufio.NewWriter(zip_file_obj)

	// Connecting to Workspace Owner Again
	// Now Transfer Data using KCP ONLY, No RPC in chunks
	kcp_conn, err := kcp.DialWithConnAndOptions(workspace_owner_ip, nil, 0, 0, udp_conn)
	if err != nil {
		fmt.Println("Error while Dialing Workspace Owner to Get Data:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	defer kcp_conn.Close()

	// KCP Params for Congestion Control
	kcp_conn.SetWindowSize(128, 1024)
	kcp_conn.SetNoDelay(1, 10, 2, 1)
	kcp_conn.SetACKNoDelay(true)
	kcp_conn.SetDSCP(46)

	// Sending the Type of Session
	kpc_buff := [3]byte{'K', 'C', 'P'}
	_, err = kcp_conn.Write(kpc_buff[:])
	if err != nil {
		fmt.Println("Error while Writing the type of Session(KCP-RPC or KCP-Plain):", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// Sending Workspace Name & Hash
	_, err = kcp_conn.Write([]byte(workspace_name))
	if err != nil {
		fmt.Println("Error while Sending Workspace Name to Workspace Owner:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	_, err = kcp_conn.Write([]byte(res.RequestPushRange))
	if err != nil {
		fmt.Println("Error while Sending Workspace Name to Workspace Owner:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// Sending Get Data Type (Pull/Clone)
	_, err = kcp_conn.Write([]byte("Clone"))
	if err != nil {
		fmt.Println("Error while Sending 'Clone' to Workspace Owner:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	buffer := make([]byte, DATA_CHUNK)

	fmt.Println("Len Data Bytes:", res.LenData)
	offset := 0

	for offset < res.LenData {
		n, err := kcp_conn.Read(buffer)
		if err != nil {
			fmt.Println("\nError while Reading from Workspace Owner:", err)
			fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
			return err
		}

		// Check for Errors on Workspace Owner's Side
		if n < 30 {
			msg := string(buffer[:n])
			if msg == "Incorrect Workspace Name/Push Num" || msg == "Incorrect Workspace Name/Push Num Range" || msg == "Internal Server Error" {
				fmt.Println("\nError while Reading from Workspace on his/her side:", msg)
				fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
				return errors.New(msg)
			}
		}

		// Decrypt Data
		decrypted_data, err := encrypt.EncryptDecryptChunk(buffer[:n], []byte(key), []byte(iv))
		if err != nil {
			fmt.Println("Error while Decrypting Chunk:", err)
			fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
			return err
		}

		// Store data in chunks using 'writer'
		_, err = writer.Write(decrypted_data)
		if err != nil {
			fmt.Println("Error while Writing Decrypted Data in Chunks:", err)
			fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
			return err
		}

		// Flush buffer to disk after 'FLUSH_AFTER_EVERY_X_CHUNK'
		if offset%FLUSH_AFTER_EVERY_X_MB == 0 {
			err = writer.Flush()
			if err != nil {
				fmt.Println("Error flushing 'writer' after X KB/MB buffer:", err)
				fmt.Println("Soure: fetchAndStoreDataIntoWorkspace()")
				return err
			}
		}

		offset += n
		utils.PrintProgressBar(offset, res.LenData, 100)
	}
	fmt.Println("\nData Transfer Completed ...")

	// Flush buffer to disk at the end
	err = writer.Flush()
	if err != nil {
		fmt.Println("Error flushing 'writer' buffer:", err)
		fmt.Println("Soure: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	zip_file_obj.Close()

	_, err = kcp_conn.Write([]byte("Data Received"))
	if err != nil {
		fmt.Println("Error while Sending Data Received Message:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		// Not Returning Error because, we got data, we don't care if workspace owner now is offline or not responding
	}

	if err = filetracker.CleanFilesFromWorkspace(workspace_path); err != nil {
		fmt.Println("Error while Cleaning Workspace :", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// Unzip Content
	if err = filetracker.UnzipData(zip_file_path, workspace_path); err != nil {
		fmt.Println("Error while Unzipping Data into Workspace:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}

	// Remove Zip File After Unzipping it
	err = os.Remove(zip_file_path)
	if err != nil {
		fmt.Println("Error while Removing the Zip File After Use:", err)
		fmt.Println("Source: fetchAndStoreDataIntoWorkspace()")
		return err
	}
	return nil
}

func Clone(workspace_owner_username, workspace_name, workspace_password string) {
	// Create .PKr folder to store zipped data
	curr_dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error while Getting Current Directory:", err)
		fmt.Println("Source: Clone()")
		return
	}

	err = os.Mkdir(filepath.Join(curr_dir, ".PKr"), 0700)
	if err != nil {
		if os.IsExist(err) {
			fmt.Println("It seems you've already used Clone in this Directory")
			return
		}
		fmt.Println("Error while using Mkdir for '.PKr' folder:", err)
		fmt.Println("Source: Clone()")
		return
	}

	err = os.Mkdir(filepath.Join(curr_dir, ".PKr", "Contents"), 0700)
	if err != nil {
		fmt.Println("Error while using Mkdir for '.PKr/Contents' folder:", err)
		fmt.Println("Source: Clone()")
		return
	}

	// Get Details from user-config
	user_conf, err := config.ReadFromUserConfigFile()
	if err != nil {
		fmt.Println("Error while Reading user-config:", err)
		fmt.Println("Source: Clone()")
		return
	}

	// Connecting to Workspace Owner
	client_handler_name, workspace_owner_ip, udp_conn, kcp_conn, err := connectToAnotherUser(workspace_name, workspace_owner_username, user_conf.ServerIP, user_conf.Username, user_conf.Password)
	if err != nil {
		fmt.Println("Error while Connecting to Another User:", err)
		fmt.Println("Source: Clone()")
		return
	}
	defer udp_conn.Close()
	defer kcp_conn.Close()

	// Sending the Type of Session
	rpc_buff := [3]byte{'R', 'P', 'C'}
	_, err = kcp_conn.Write(rpc_buff[:])
	if err != nil {
		fmt.Println("Error while Writing the type of Session(KCP-RPC or KCP-Plain):", err)
		fmt.Println("Source: cloneWorkspace()")
		return
	}

	// Creating RPC Client
	rpc_client := rpc.NewClient(kcp_conn)
	defer rpc_client.Close()
	rpcClientHandler := dialer.ClientCallHandler{}

	// Get Public Key of Workspace Owner
	fmt.Println("Requesting Public Key of Workspace Owner ...")
	public_key, err := rpcClientHandler.CallGetPublicKey(client_handler_name, rpc_client)
	if err != nil {
		fmt.Println("Error while Calling GetPublicKey:", err)
		fmt.Println("Source: Clone()")
		return
	}

	// Store the Public Key of Workspace Owner
	fmt.Println("Storing Public Key of Workspace Owner ...")
	err = config.StorePublicKeyOfOtherUser(workspace_owner_username, public_key)
	if err != nil {
		fmt.Println("Error while Storing the Public Key of Workspace Owner:", err)
		fmt.Println("Source: Clone()")
		return
	}

	// Encrypting Workspace Password with Public Key
	fmt.Println("Encrypting  Workspace Password with Public Key ...")
	encrypted_password, err := encrypt.RSAEncryptData(workspace_password, string(public_key))
	if err != nil {
		fmt.Println("Error while Encrypting Workspace Password via Public Key:", err)
		fmt.Println("Source: Clone()")
		return
	}

	// Reading My Public Key
	my_public_key, err := config.ReadMyPublicKey()
	if err != nil {
		fmt.Println("Error while Reading Public Key:", err)
		fmt.Println("Source: Clone()")
		return
	}
	base64_public_key := []byte(base64.StdEncoding.EncodeToString(my_public_key))

	// Requesting InitWorkspaceConnection
	fmt.Println("Calling Requesting InitWorkspaceConnection ...")
	err = rpcClientHandler.CallInitNewWorkSpaceConnection(workspace_name, user_conf.Username, user_conf.ServerIP, encrypted_password, base64_public_key, client_handler_name, rpc_client)
	if err != nil {
		fmt.Println("Error while Calling Init New Workspace Connection:", err)
		fmt.Println("Source: Clone()")
		return
	}

	fmt.Println("Requesting Meta Data ...")
	// Calling GetMetaData
	res, err := rpcClientHandler.CallGetMetaData(user_conf.Username, user_conf.ServerIP, workspace_name, encrypted_password, client_handler_name, -1, rpc_client)
	if err != nil {
		fmt.Println("Error while Calling GetMetaData:", err)
		fmt.Println("Source: Clone()")
		return
	}

	kcp_conn.Close()
	rpc_client.Close()

	err = fetchAndStoreDataIntoWorkspace(workspace_owner_ip, workspace_name, udp_conn, *res)
	if err != nil {
		fmt.Println("Error while Fetching & Storing Data:", err)
		fmt.Println("Source: Clone()")
		return
	}
	fmt.Println("Data Stored into Workspace ...")

	// Register New User to Workspace
	// New GRPC Client
	fmt.Println("Sending Updates to Server ...")
	gRPC_cli_service_client, err := dialer.GetNewGRPCClient(user_conf.ServerIP)
	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Description: Cannot Create New GRPC Client")
		fmt.Println("Source: Clone()")
		return
	}

	// Prepare req
	register_user_to_workspace_res_req := &pb.RegisterUserToWorkspaceRequest{
		ListenerUsername:       user_conf.Username,
		ListenerPassword:       user_conf.Password,
		WorkspaceName:          workspace_name,
		WorkspaceOwnerUsername: workspace_owner_username,
	}

	// Request Timeout
	ctx, cancelFunc := context.WithTimeout(context.Background(), CONTEXT_TIMEOUT)
	defer cancelFunc()

	// Sending Request to Server
	_, err = gRPC_cli_service_client.RegisterUserToWorkspace(ctx, register_user_to_workspace_res_req)
	if err != nil {
		fmt.Println("Error while Registering User To Workspace:", err)
		fmt.Println("Source: Clone()")
		return
	}

	// Update tmp/userConfig.json
	err = config.RegisterNewGetWorkspace(workspace_name, workspace_owner_username, curr_dir, workspace_password, res.LastPushNum)
	if err != nil {
		fmt.Println("Error while Registering New GetWorkspace:", err)
		fmt.Println("Source: Clone()")
		return
	}
	fmt.Println("Workspace Cloned Successfully")
}
