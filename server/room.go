package server

import (
	"fmt"
	"sync"
	"time"

	"GO-chat-simulator/models"
)

type Room struct {
	name    string
	clients map[*Client]bool
	mutex   sync.RWMutex
}

func NewRoom(name string) *Room {
	return &Room{
		name:    name,
		clients: make(map[*Client]bool),
	}
}

func (r *Room) addClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.clients[client] = true
}

func (r *Room) removeClient(client *Client) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	delete(r.clients, client)
}

func (r *Room) broadcast(message models.Message, exclude *Client) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	timestamp := time.Now().Format("15:04")
	var formattedMessage string

	switch message.Type {
	case models.ChatMessage:
		formattedMessage = fmt.Sprintf("[%s] %s: %s", timestamp, message.Username, message.Content)
	case models.SystemMessage:
		formattedMessage = fmt.Sprintf("[%s] %s", timestamp, message.Content)
	}

	for client := range r.clients {
		if exclude != nil && client == exclude {
			continue
		}
		client.send(formattedMessage)
	}
}

func (r *Room) getUserCount() int {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return len(r.clients)
}

func (r *Room) getUsers() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	users := make([]string, 0, len(r.clients))
	for client := range r.clients {
		users = append(users, client.username)
	}
	return users
}
