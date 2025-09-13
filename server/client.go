package server //GO-chat-simulator/server

import (
	"fmt"
	"net"
)

type Client struct {
	conn     net.Conn
	username string
	room     *Room
}

func (c *Client) send(message string) {
	_, err := c.conn.Write([]byte(message + "\n"))
	if err != nil {
		fmt.Printf("Error sending message to %s: %v\n", c.username, err)
	}
}
