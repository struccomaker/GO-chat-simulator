package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"

	"GO-chat-simulator/models"
)

type Server struct {
	rooms        map[string]*Room
	clients      map[net.Conn]*Client
	lastMessage  *models.Message     // Track last message globally
	lastSender   *Client             // Track who sent the last message
	replyTargets map[*Client]*Client // Track who each client should reply to
	mutex        sync.RWMutex
}

func NewServer() *Server {
	server := &Server{
		rooms:        make(map[string]*Room),
		clients:      make(map[net.Conn]*Client),
		replyTargets: make(map[*Client]*Client),
	}

	// create default rooms
	server.createRoom("general")
	server.createRoom("random")
	server.createRoom("tech")

	return server
}

func (s *Server) HandleConnection(conn net.Conn) {
	defer conn.Close()

	// create new client with temporary username
	client := &Client{
		conn:     conn,
		username: fmt.Sprintf("User_%d", len(s.clients)+1),
		room:     nil,
	}

	s.mutex.Lock()
	s.clients[conn] = client
	s.mutex.Unlock()

	fmt.Printf("New client connected: %s\n", client.username)

	// handle client messages
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			continue
		}

		s.processCommand(client, message)
	}

	// client disconnected
	s.removeClient(client)
	fmt.Printf("Client disconnected: %s\n", client.username)
}

func (s *Server) processCommand(client *Client, input string) {
	if !strings.HasPrefix(input, "/") {
		client.send("Commands must start with /. Try /list, /join <room>, /msg <username> <message>, /all <message>, /global <message>, or /r <reply>")
		return
	}

	parts := strings.SplitN(input, " ", 2)
	command := parts[0]

	switch command {
	case "/setname":
		if len(parts) < 2 {
			client.send("Usage: /setname <username>")
			return
		}
		s.setUsername(client, parts[1])

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
			client.send("Usage: /msg <username> <message>")
			return
		}
		s.sendPrivateMessage(client, parts[1])

	case "/all":
		if len(parts) < 2 {
			client.send("Usage: /all <your_message>")
			return
		}
		s.sendToRoom(client, parts[1])

	case "/global":
		if len(parts) < 2 {
			client.send("Usage: /global <your_message>")
			return
		}
		s.broadcastToAll(client, parts[1])

	case "/r":
		if len(parts) < 2 {
			client.send("Usage: /r <reply_message>")
			return
		}
		s.replyToLast(client, parts[1])

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
		client.send("Unknown command. Available: /list, /join <room>, /msg <username> <message>, /all <message>, /global <message>, /r <reply>, /users, /create <room>")
	}
}

func (s *Server) setUsername(client *Client, username string) {
	// Trim whitespace and validate username
	username = strings.TrimSpace(username)
	if username == "" {
		client.send("Username cannot be empty or just whitespace!")
		return
	}

	// Check if username contains spaces
	if strings.Contains(username, " ") {
		client.send("Username cannot contain spaces! Please choose a different username.")
		return
	}

	// Check if username is already taken
	s.mutex.RLock()
	for _, c := range s.clients {
		if c.username == username && c != client {
			s.mutex.RUnlock()
			client.send("Username already taken! Please choose a different username.")
			return
		}
	}
	s.mutex.RUnlock()

	oldName := client.username
	client.username = username

	welcome := fmt.Sprintf("Welcome %s! Type /list to see available rooms.", client.username)
	client.send(welcome)

	fmt.Printf("Client %s changed name to: %s\n", oldName, username)
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
		client.send(fmt.Sprintf("  - %s (%d users)", name, userCount))
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

	// leave current room if in one
	if client.room != nil {
		client.room.removeClient(client)
		client.room.broadcast(models.Message{
			Username: "System",
			Content:  fmt.Sprintf("%s left the room", client.username),
			Type:     models.SystemMessage,
		}, nil)
	}

	// join new room (create if doesn't exist)
	room, exists := s.rooms[roomName]
	if !exists {
		room = NewRoom(roomName)
		s.rooms[roomName] = room
	}

	room.addClient(client)
	client.room = room

	client.send(fmt.Sprintf("Joined room: %s", roomName))

	// announce to room
	room.broadcast(models.Message{
		Username: "System",
		Content:  fmt.Sprintf("%s joined the room", client.username),
		Type:     models.SystemMessage,
	}, client) // exclude the joining client
}

func (s *Server) sendToRoom(client *Client, content string) {
	if client.room == nil {
		client.send("You must join a room first! Use /list to see available rooms.")
		return
	}

	message := models.Message{
		Username: client.username,
		Content:  content,
		Type:     models.ChatMessage,
	}

	// store as last message globally
	s.mutex.Lock()
	s.lastMessage = &message
	s.lastSender = client
	// Set reply target for all clients in the room to this sender
	for clientInRoom := range client.room.clients {
		if clientInRoom != client {
			s.replyTargets[clientInRoom] = client
		}
	}
	s.mutex.Unlock()

	// broadcast to room (this will show timestamp and proper formatting)
	client.room.broadcast(message, nil)
}

func (s *Server) broadcastToAll(client *Client, content string) {
	message := models.Message{
		Username: client.username,
		Content:  content,
		Type:     models.ChatMessage,
	}

	// store as last message globally and set reply targets
	s.mutex.Lock()
	s.lastMessage = &message
	s.lastSender = client
	// set reply target for all clients to this sender
	for _, c := range s.clients {
		if c != client {
			s.replyTargets[c] = client
		}
	}
	s.mutex.Unlock()

	// broadcast to all connected clients
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	for _, c := range s.clients {
		if c != client { // exclude sender
			c.send(fmt.Sprintf("[GLOBAL] %s: %s", client.username, content))
		}
	}

	// confirm to sender
	client.send(fmt.Sprintf("[GLOBAL MESSAGE SENT]: %s", content))
}

func (s *Server) sendPrivateMessage(client *Client, input string) {
	// parse: username message content
	parts := strings.SplitN(input, " ", 2)
	if len(parts) < 2 {
		client.send("Usage: /msg <username> <message>")
		return
	}

	targetUsername := parts[0]
	messageContent := parts[1]

	// find target client
	s.mutex.RLock()
	var targetClient *Client
	for _, c := range s.clients {
		if c.username == targetUsername {
			targetClient = c
			break
		}
	}
	s.mutex.RUnlock()

	if targetClient == nil {
		client.send(fmt.Sprintf("User '%s' not found or not online", targetUsername))
		return
	}

	if targetClient == client {
		client.send("You cannot send a private message to yourself!")
		return
	}

	// create message for last message tracking
	message := models.Message{
		Username: client.username,
		Content:  fmt.Sprintf("[Private to %s] %s", targetUsername, messageContent),
		Type:     models.ChatMessage,
	}

	// store as last message and set reply targets
	s.mutex.Lock()
	s.lastMessage = &message
	s.lastSender = client
	// set reply targets: target can reply to sender, sender can reply to target
	s.replyTargets[targetClient] = client
	s.replyTargets[client] = targetClient
	s.mutex.Unlock()

	// send private message
	privateMsg := fmt.Sprintf("[Private from %s]: %s", client.username, messageContent)
	targetClient.send(privateMsg)

	// confirmation to sender
	client.send(fmt.Sprintf("[Private to %s]: %s", targetUsername, messageContent))
}

func (s *Server) replyToLast(client *Client, reply string) {
	s.mutex.RLock()
	replyTarget := s.replyTargets[client]
	s.mutex.RUnlock()

	if replyTarget == nil {
		client.send("No one to reply to. Send or receive a message first.")
		return
	}

	// check if target is still connected
	s.mutex.RLock()
	targetExists := false
	for _, c := range s.clients {
		if c == replyTarget {
			targetExists = true
			break
		}
	}
	s.mutex.RUnlock()

	if !targetExists {
		client.send(fmt.Sprintf("Cannot reply - %s is no longer connected.", replyTarget.username))
		// clean up the reply target
		s.mutex.Lock() //LOCK HERE!!
		delete(s.replyTargets, client)
		s.mutex.Unlock()
		return
	}

	// create reply message
	replyContent := fmt.Sprintf("@%s %s", replyTarget.username, reply)

	// Send reply only to the target
	replyMsg := fmt.Sprintf("[REPLY from %s]: %s", client.username, replyContent)
	replyTarget.send(replyMsg)

	// cfm  to replier
	client.send(fmt.Sprintf("[REPLY to %s]: %s", replyTarget.username, replyContent))

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
			client.send(fmt.Sprintf("  - %s (you)", username))
		} else {
			client.send(fmt.Sprintf("  - %s", username))
		}
	}
}

func (s *Server) removeClient(client *Client) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// remove from room
	if client.room != nil {
		client.room.removeClient(client)
		client.room.broadcast(models.Message{
			Username: "System",
			Content:  fmt.Sprintf("%s left the room", client.username),
			Type:     models.SystemMessage,
		}, nil)
	}

	// clean up reply targets
	delete(s.replyTargets, client)
	// Also remove this client as a reply target for others
	for c, target := range s.replyTargets {
		if target == client {
			delete(s.replyTargets, c)
		}
	}

	// remove from server
	delete(s.clients, client.conn)
}
