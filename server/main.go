package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
)

var port string

func createServerAcceptConnectionsSocket() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func AcceptConnectionsFromClients(listener net.Listener) error {
	for {
		//This is a blocking call that will wait for a client to connect
		conn, err := listener.Accept()
		if err != nil {
			// Check if the error is due to listener being closed (expected during shutdown)
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Println("Listener closed, shutting down accept loop")
				return nil
			}
			log.Println("Error accepting connection: ", err)
			return err
		}

		fmt.Println("New connection from client: ", conn.RemoteAddr())
	}
}

func main() {
	// Load environment variables from .env file in parent directory
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: .env file not found in parent directory")
	}

	port = os.Getenv("SERVER_CONNECTIONS_PORT")
	if port == "" {
		log.Println("Port not found in environment file, Using default port 7734")
		port = "7734"
	}

	//Create a TCP socket to accept connections from clients
	listener,err := createServerAcceptConnectionsSocket()
	if err != nil {
		log.Fatal("Error creating server accept connections socket: ", err)
	}

	log.Println("Server is running on port ", port)

	//Accept connections from clients in a separate goroutine
	go func() {
		err = AcceptConnectionsFromClients(listener)
		if err != nil {
			log.Println("Error in accept connections goroutine: ", err)
		}
	}()

	// We need to block the main go routine to keep the server running
	//Main thread will exit on user interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Server is shutting down...")
	listener.Close()
}