package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Server struct {
	rooms   map[string]*Room
	clients map[net.Conn]*Client
	mutex   sync.RWMutex
}

func NewServer() *Server {
	server := &Server{
		rooms:   make(map[string]*Room),
		clients: make(map[net.Conn]*Client),
	}

	// Create default rooms
	server.createRoom("general")
	server.createRoom("random")
	server.createRoom("tech")

	return server
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// Create new client
	client := &Client{
		conn:     conn,
		username: fmt.Sprintf("User_%d", len(s.clients)+1),
		room:     nil,
	}

	s.mutex.Lock()
	s.clients[conn] = client
	s.mutex.Unlock()

	fmt.Printf("New client connected: %s\n", client.username)

	// Send welcome message
	welcome := fmt.Sprintf("Welcome %s! Type /list to see available rooms.", client.username)
	client.send(welcome)

	// Handle client messages
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		s.processCommand(client, message)
	}

	// Client disconnected
	s.removeClient(client)
	fmt.Printf("Client disconnected: %s\n", client.username)
}

func (s *Server) processCommand(client *Client, input string) {
	if !strings.HasPrefix(input, "/") {
		client.send("Commands must start with /. Try /list, /join <room>, or /msg <message>")
		return
	}

	parts := strings.SplitN(input, " ", 2)
	command := parts[0]

	switch command {
	case "/list":
		s.listRooms(client)

	case "/join":
		if len(parts) < 2 {
			client.send("Usage: /join <room_name>")
			return
		}
		s.joinRoom(client, parts[1])

	case "/msg":
		if len(parts) < 2 {
			client.send("Usage: /msg <your_message>")
			return
		}
		s.sendMessage(client, parts[1])

	case "/users":
		s.listUsersInRoom(client)

	case "/create":
		if len(parts) < 2 {
			client.send("Usage: /create <room_name>")
			return
		}
		s.createRoom(parts[1])
		client.send(fmt.Sprintf("Room '%s' created! Use /join %s to enter.", parts[1], parts[1]))

	default:
		client.send("Unknown command. Available: /list, /join <room>, /msg <message>, /users, /create <room>")
	}
}

func (s *Server) listRooms(client *Client) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.rooms) == 0 {
		client.send("No rooms available")
		return
	}

	client.send("Available rooms:")
	for name, room := range s.rooms {
		userCount := room.getUserCount()
		client.send(fmt.Sprintf("   • %s (%d users)", name, userCount))
	}
}

func (s *Server) createRoom(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, exists := s.rooms[name]; !exists {
		s.rooms[name] = NewRoom(name)
	}
}

func (s *Server) joinRoom(client *Client, roomName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Leave current room if in one
	if client.room != nil {
		client.room.removeClient(client)
		client.room.broadcast(models.Message{
			Username: "System",
			Content:  fmt.Sprintf("%s left the room", client.username),
			Type:     models.SystemMessage,
		}, nil)
	}

	// Join new room (create if doesn't exist)
	room, exists := s.rooms[roomName]
	if !exists {
		room = NewRoom(roomName)
		s.rooms[roomName] = room
	}

	room.addClient(client)
	client.room = room

	client.send(fmt.Sprintf("Joined room: %s", roomName))

	// Announce to room
	room.broadcast(models.Message{
		Username: "System",
		Content:  fmt.Sprintf("%s joined the room", client.username),
		Type:     models.SystemMessage,
	}, client) // Exclude the joining client
}

func (s *Server) sendMessage(client *Client, content string) {
	if client.room == nil {
		client.send("You must join a room first! Use /list to see available rooms.")
		return
	}

	message := models.Message{
		Username: client.username,
		Content:  content,
		Type:     models.ChatMessage,
	}

	client.room.broadcast(message, nil)
}

func (s *Server) listUsersInRoom(client *Client) {
	if client.room == nil {
		client.send("You must join a room first!")
		return
	}

	users := client.room.getUsers()
	if len(users) == 0 {
		client.send("No users in this room")
		return
	}

	client.send(fmt.Sprintf("Users in %s:", client.room.name))
	for _, username := range users {
		if username == client.username {
			client.send(fmt.Sprintf("   • %s (you)", username))
		} else {
			client.send(fmt.Sprintf("   • %s", username))
		}
	}
}

func (s *Server) removeClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Remove from room
	if client.room != nil {
		client.room.removeClient(client)
		client.room.broadcast(models.Message{
			Username: "System",
			Content:  fmt.Sprintf("%s left the room", client.username),
			Type:     models.SystemMessage,
		}, nil)
	}

	// Remove from server
	delete(s.clients, client.conn)
}
