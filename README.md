# 🍻 PKr
PKr is a peer-to-peer workspace sharing and synchronization tool written in Go.

It allows users to share, clone, and push workspaces directly between machines — with encryption — through a central coordination server but without routing file data through it.

### Q. Why should you use ?
- In situations for 1 sender- N reciever data transfer over the internet with automatic synchronization.

   Example: Automatically Sharing Homeworks & Practicals with your classmates 😀

> [!NOTE]
> - Workspaces : Folders/Directories are refered to as workspaces.
> - GetWorkspaces  : These are the cloned workspaces, where you are the receiver.
> - SendWorkspaces : These are the workspaces where you are the sender, you will be pushing updates.

## 1. Features
- [X] Encrypted Data Transfer via RSA and AES.
- [X] P2P Data transfer over the internet via NAT Punching.
- [X] Push Changes made to a worksapace
- [X] Automatic Synchronization of Workspaces.
- [X] Provides Notification for New Updates to Receiver.

### TODO
- [ ] Test the P2P reliability on different ISPs
- [ ] Simplify the Punching and Connection Establishment Logic
- [ ] Verify provided changes via hash during PUSH/PULL

> [!IMPORTANT]
> For troubleshooting and monitoring, logs are stored in :
>  - Windows : %APPDATA%/Local/PKr/Logs
>  - Linux : ~/.local/share/PKr/Logs

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

> [!WARNING]
> - Run the command in an **Empty** directory, this command will remove all the files and folders of this directory

> [!IMPORTANT]
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

```
PKr-Cli.exe push
```

Required Fields:
   - Workspace Name (Read the init **Note** to learn More)
   - Push Description (This will be visible to receiving user's notification)

#### Senders Side
![sender's push command ss1](/assets/push_sender.png "PKr Push Sender POV")

#### Receiver's Side

> [!NOTE]
> - The changes are automatically reflected
> - The Receiver is notified of changes via Notification

![receivers's side ss1](/assets/push_receiver.png "PKr Push Receiver POV")

**Notification**
![receivers's side ss2](/assets/push_receiver_notify.png "PKr Push Receiver Notification POV")

### 3.4 List all Workspace
Lists all workspaces - GetWorkspaces & SendWorkspaces
```
PKr-Cli.exe list
```




## 4. Bugs & Issues
For any bugs & issues you can post about it in the **issues** with the Log file & steps to reproduce it.