# GO Chat Simulator

A real-time TCP-based chat server and client application built in Go within CMD. Supports multiple chat rooms, private messaging and global broadcasts. Built for learning.
## Features

- **Multi-room Chat**: Create and join different chat rooms
- **Private Messaging**: Send direct messages to specific users
- **Global Broadcasting**: Send messages to all connected users
- **Real-time Communication**: TCP-based instant messaging
- **Thread-safe**: Concurrent client handling with proper synchronization

## Project Structure

```
GO-chat-simulator/
├── main.go           # Entry point (server/client launcher)
├── models/
│   └── message.go    # Message types and structures
├── server/
│   ├── server.go     # Core server logic and command processing
│   ├── room.go       # Room management and broadcasting
│   └── client.go     # Client connection handling
├── go.mod           # Go module definition
└── README.md        # This file
```

## Installation & Setup

1. **Clone or download the project files**

2. **Initialize the Go module** (if not already done):
   ```bash
   go mod init GO-chat-simulator
   ```

3. **Build the application**:
   ```bash
   go build
   ```

## Usage

### Starting the Server

```bash
go run main.go
```

The server will prompt you to enter a port number (default: 8080).

### Connecting as a Client

```bash
go run main.go client
```



## Commands

### Basic Commands
- `/list` - Show available chat rooms
- `/join <room>` - Join a specific chat room
- `/create <room>` - Create a new chat room
- `/users` - List users in your current room
- `/quit` - Disconnect from the server

### Messaging Commands
- `/all <message>` - Send message to all users in your current room
- `/global <message>` - Send message to ALL users across all rooms
- `/msg <username> <message>` - Send private message to a specific user
- `/r <reply>` - Reply to the last person who sent you a message

### Username Management
- `/setname <username>` - Change your username


## Technical Details

- **Language**: Go 1.21+
- **Protocol**: TCP
- **Concurrency**: Goroutines for each client connection
- **Synchronization**: RWMutex for thread-safe operations
- **Architecture**: Client-server model with room-based organization


## License

This project is open source and available under the [MIT License](LICENSE).