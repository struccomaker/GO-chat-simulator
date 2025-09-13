package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"GO-chat-simulator/server"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "client" {
		runClient()
		return
	}
	
	runServer()
}

func runServer() {
	chatServer := server.NewServer()
	
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	fmt.Println("Chat Server started on port 8080")
	fmt.Println("Waiting for connections...")
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Failed to accept connection:", err)
			continue
		}
		
		go chatServer.HandleConnection(conn)
	}
}

func runClient() {
	// get server address and port
	fmt.Print("Enter server address and port (default localhost:8080): ")
	scanner := bufio.NewScanner(os.Stdin)
	address := "localhost:8080" // default
	
	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			address = input
		}
	}
	
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	// get username from user
	fmt.Print("Enter your username: ")
	if !scanner.Scan() {
		log.Fatal("Failed to read username")
	}
	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		username = "Anonymous"
	}
	
	// send username to server
	_, err = conn.Write([]byte("/setname " + username + "\n"))
	if err != nil {
		log.Fatal("Failed to send username:", err)
	}

	fmt.Printf("Connected to chat server as %s!\n", username)
	fmt.Println("Commands: /list, /join <room>, /msg <username> <message>, /r <reply>, /quit")
	fmt.Println("Start by typing /list to see available rooms")
	fmt.Println()

	// start goroutine to read messages from server
	go func() {
		serverScanner := bufio.NewScanner(conn)
		for serverScanner.Scan() {
			fmt.Println(serverScanner.Text())
		}
	}()

	// read input from user and send to server
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		
		input := strings.TrimSpace(scanner.Text())
		if input == "/quit" {
			break
		}
		
		if input == "" {
			continue
		}
		
		_, err := conn.Write([]byte(input + "\n"))
		if err != nil {
			log.Println("Error sending message:", err)
			break
		}
	}
	
	fmt.Println("Disconnected from server")
}