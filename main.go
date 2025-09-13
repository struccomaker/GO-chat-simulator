package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
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
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Failed to connect to server:", err)
	}
	defer conn.Close()

	fmt.Println("Connected to chat server!")
	fmt.Println("Commands: /list, /join <room>, /msg <message>, /quit")
	fmt.Println("Start by typing /list to see available rooms")
	fmt.Println()

	// Start goroutine to read messages from server
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// Read input from user and send to server
	scanner := bufio.NewScanner(os.Stdin)
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