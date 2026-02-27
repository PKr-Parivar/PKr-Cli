package main

import (
	"bufio"
	"fmt"

	"os"

	"strings"

	"github.com/PKr-Parivar/PKr-Base/utils"
	"github.com/PKr-Parivar/PKr-Cli/root"
	"github.com/pkg/profile"
	// _ "net/http/pprof" // Temp Profile Import - Remove Later
)

var (
	IS_PROF      = false
	PROFILE      interface{ Stop() }
	PROFILE_ARGS []func(*profile.Profile)
)

func printArguments() {
	fmt.Println("Valid Parameters:")
	fmt.Println("	1] install -> Create User and Install PKr")
	fmt.Println("	2] init -> Initialize a Workspace, allows other Users to connect")
	fmt.Println("	3] clone -> Clone an existing Workspace of a different User")
	fmt.Println("	4] list -> List all Workspaces")
	fmt.Println("	5] push -> Push new Changes to Listeners")
	fmt.Println("	6] debug -> Add Additional cmd line args")
}

func main() {
	if len(os.Args) < 2 {
		printArguments()
		return
	}

	cmd := make([]string, len(os.Args))
	for i := 0; i < len(os.Args); i++ {
		cmd[i] = strings.ToLower(os.Args[i])
	}

	INDEX := 1
	if cmd[INDEX] == "debug" {
		fmt.Println("[DEBUG MODE is ON]")
		if len(os.Args) > 2 && os.Args[1] == "debug" {
			for i, arg := range os.Args {
				if arg == "--fp" && i+1 < len(os.Args) {
					fmt.Println("[DEBUG] Setting Debug Config File Path: ", os.Args[i+1])
					utils.SetUserConfigDir(os.Args[i+1])

					INDEX = i + 2
				} else if arg == "--prof.cpu" && i+1 < len(os.Args) {
					IS_PROF = true

					fmt.Println("[DEBUG] Generating CPU Profiles at: ", os.Args[i+1])

					PROFILE_ARGS = append(PROFILE_ARGS, profile.CPUProfile)
					PROFILE_ARGS = append(PROFILE_ARGS, profile.ProfilePath(os.Args[i+1]))

					INDEX = i + 2
				} else if arg == "--prof.mem" && i+1 < len(os.Args) {
					IS_PROF = true

					fmt.Println("[DEBUG] Generating CPU Profiles at: ", os.Args[i+1])

					PROFILE_ARGS = append(PROFILE_ARGS, profile.MemProfile)
					PROFILE_ARGS = append(PROFILE_ARGS, profile.MemProfileRate(1))
					PROFILE_ARGS = append(PROFILE_ARGS, profile.ProfilePath(os.Args[i+1]))

					INDEX = i + 2
				} else if arg == "--prof.trace" && i+1 < len(os.Args) {
					IS_PROF = true

					fmt.Println("[DEBUG] Generating CPU Profiles at: ", os.Args[i+1])
					PROFILE_ARGS = append(PROFILE_ARGS, profile.TraceProfile)
					PROFILE_ARGS = append(PROFILE_ARGS, profile.ProfilePath(os.Args[i+1]))

					INDEX = i + 2
				}

			}
		}
	}

	// Set Command Line Arguments
	// Current :
	// 		1. --cpath=<file-path>  ;;  Set config-path when executing cmd

	for i := INDEX + 1; i < len(os.Args)-INDEX + 1; i++ {
		if strings.Contains(cmd[i], "--cpath") {
			out := strings.Split(cmd[i], "=")
			if len(out) != 2 {
				fmt.Println("Error: --cpath values is not properly provided")
				fmt.Println("example: --cpath=<filepath>")
				return
			}

			err := utils.SetUserConfigDir(out[1])
			if err != nil {
				fmt.Printf("Error while Setting user-config to %s: %v\n", out[1], err)
				fmt.Println("Source: Install()")
				return
			}

			fmt.Println("[DEBUG] install_path is set: ", out)
			fmt.Println("[DEBUG] !! Keep in Mind !! This is where the services will look for Config File")
		}
	}

	switch cmd[INDEX] {
	case "install":
		{
			var server_ip, username, password string

			fmt.Print("> Enter Username: ")
			fmt.Scan(&username)

			fmt.Print("> Enter Password: ")
			fmt.Scan(&password)

			fmt.Print("> Enter Server IP: ")
			fmt.Scan(&server_ip)

			if IS_PROF {
				fmt.Println("Starting Profiling....")
				PROFILE = profile.Start(PROFILE_ARGS...)
			}
			root.Install(server_ip, username, password)
		}

	case "init":
		{
			var workspace_password, push_desc string
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("> Enter Workspace Password: ")
			workspace_password, _ = reader.ReadString('\n')
			workspace_password = strings.TrimSpace(workspace_password)

			fmt.Print("> Enter Push Description: ")
			push_desc, _ = reader.ReadString('\n')
			push_desc = strings.TrimSpace(push_desc)

			if IS_PROF {
				fmt.Println("Starting Profiling....")
				PROFILE = profile.Start(PROFILE_ARGS...)
			}
			root.InitWorkspace(workspace_password, push_desc)
		}

	case "clone":
		{
			var workspace_owner_username string
			var workspace_name string
			var workspace_password string

			fmt.Println("WARNING: All Previous files'll be DELETED & REPLACED by files Received from Workspace Owner")
			fmt.Print("> Enter the Workspace Owner Username: ")
			fmt.Scan(&workspace_owner_username)

			fmt.Print("> Enter Workspace Name: ")
			fmt.Scan(&workspace_name)

			fmt.Print("> Enter Workspace Password: ")
			fmt.Scan(&workspace_password)

			fmt.Println("Cloning ...")

			if IS_PROF {
				fmt.Println("Starting Profiling....")
				PROFILE = profile.Start(PROFILE_ARGS...)
			}
			root.Clone(workspace_owner_username, workspace_name, workspace_password)
		}

	case "list":
		{
			fmt.Println("Fetching All Workspaces from Server ...")

			if IS_PROF {
				fmt.Println("Starting Profiling....")
				PROFILE = profile.Start(PROFILE_ARGS...)
			}
			root.ListAllWorkspaces()
		}

	case "push":
		{
			var workspace_name, push_desc string
			reader := bufio.NewReader(os.Stdin)

			fmt.Print("> Enter Workspace Name: ")
			workspace_name, _ = reader.ReadString('\n')
			workspace_name = strings.TrimSpace(workspace_name)

			fmt.Print("> Enter Push Description: ")
			push_desc, _ = reader.ReadString('\n')
			push_desc = strings.TrimSpace(push_desc)

			fmt.Printf("Pushing Workpace: %s ...\n", workspace_name)

			if IS_PROF {
				fmt.Println("Starting Profiling....")
				PROFILE = profile.Start(PROFILE_ARGS...)
			}
			root.Push(workspace_name, push_desc)
		}

	default:
		printArguments()
	}

	// remove before commit and push -- for deubg only
	if IS_PROF {
		fmt.Println("<<< CLOSE PROFILING >>>")
		PROFILE.Stop()
	}
}
