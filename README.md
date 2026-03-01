# PKr

## 1. Features
?? DO IT PROPERLY??
- [ ] Encrypted
- [ ] Notifications
- [ ] Manage Multiple Send & Get Workspaces

## 2. Installation

1. Download and Run the latest binary from [PKr-Service's Release](https://github.com/PKr-Parivar/PKr-Service/releases) page
2. Once Installed, Restart the PC 
3. In your Terminal: 
```bash
PKr-Cli.exe install
```
1. Fill in the details (You will need an IP for an actively running [PKr-Server](https://github.com/PKr-Parivar/PKr-Server))
   - Note: The Field `server ip` requests for the gRPC port of the PKr-Server. Eg: `server_ip:grpc_port`
2. Restart the PC one last time 🙏


## 3. Usage

### 3.1 Share your Workspace
Why? You want to share a directory/workspace with other users.

-> In the directory you want to share:
```
PKr-Cli.exe init
```
> [!NOTE]
> - Now you can make your recievers "clone" your workspace. However for it to work they need to be in the same PKr-Server.
> - Your Workspace Name is the name of the Directory from where you ran the command

Required Fields:
   - Password
   - Push Description

![init command](/assets/init.png "PKr init")

---

### 3.2 Clone Someones Workspace
This will not only initially clone the workspace, but from now own will be automatically synced.

-> In the Directory you want to clone in:
```
PKr-Cli.exe clone
```

> [!NOTE]
> - You need to be in the same PKr-Server for all of this to work.


Required Fields:
   - Owners Username
   - Workspace Name (Read the init **Note** to learn More)
   - Workspace Password

![clone command ss1](/assets/clone_1.png "PKr Clone 1")
![clone command ss2](/assets/clone_2.png "PKr Clone 2")

---

### 3.3 Update Workspace
The workspace owner can regularly send updates. These updates are provided and synced with the recievers automatically.

-> In the workspace/directory whose changes you want to push:
```
PKr-Cli.exe push
```

README IS BEING UPDATED 