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

	// Query port number from user
	fmt.Print("Enter port number to start server on (default 8080): ")
	scanner := bufio.NewScanner(os.Stdin)
	port := "8080" // default

	if scanner.Scan() {
		input := strings.TrimSpace(scanner.Text())
		if input != "" {
			port = input
		}
	}

	address := ":" + port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer listener.Close()

	fmt.Printf("Chat Server started on port %s\n", port)
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
	// auto to localhost:8080 or else get from the user
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
	var username string
	for {
		fmt.Print("Enter your username (no spaces allowed): ")
		if !scanner.Scan() {
			log.Fatal("Failed to read username")
		}
		username = strings.TrimSpace(scanner.Text())

		if username == "" {
			fmt.Println("Username cannot be empty. Please try again.")
			continue
		}

		if strings.Contains(username, " ") {
			fmt.Println("Username cannot contain spaces. Please try again.")
			continue
		}

		break // valid username
	}

	// send username to server
	_, err = conn.Write([]byte("/setname " + username + "\n"))
	if err != nil {
		log.Fatal("Failed to send username:", err)
	}

	fmt.Printf("Connected to chat server as %s!\n", username)
	fmt.Println("Commands:")
	fmt.Println("  /list - show available rooms")
	fmt.Println("  /join <room> - join a room")
	fmt.Println("  /msg <username> <message> - send private message")
	fmt.Println("  /all <message> - send message to current room")
	fmt.Println("  /global <message> - broadcast to all users everywhere")
	fmt.Println("  /r <reply> - reply privately to last message sender")
	fmt.Println("  /users - list users in current room")
	fmt.Println("  /create <room> - create new room")
	fmt.Println("  /quit - disconnect from server")
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
