package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	common_helpers "P2P/common-helpers"

	"github.com/joho/godotenv"
)

var (
	port          string
	clientCounter int

	// peerInfoMap stores mapping of hostname to upload port
	peerInfoMap      = make(map[string]string)
	peerInfoMapMutex sync.RWMutex

	// rfcIndexMap stores RFC information indexed by hostname
	// Each entry is a slice of [RFC_Number, RFC_Title] pairs
	rfcIndexMap      = make(map[string][][]string)
	rfcIndexMapMutex sync.RWMutex
)

// createServerAcceptConnectionsSocket creates the main listener for accepting client connections
func createServerAcceptConnectionsSocket() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

// handleClientConnection manages the dedicated connection for a single client
func handleClientConnection(conn net.Conn, clientID int) error {
	// Allocate a dedicated port for this client
	dedicatedPort, err := common_helpers.GetFreePort()
	if err != nil {
		log.Printf("Error getting free port for client %d: %v", clientID, err)
		return err
	}

	log.Printf("Client %d assigned dedicated port %s", clientID, dedicatedPort)

	// Create dedicated listener on the allocated port
	dedicatedListener, err := net.Listen("tcp", ":"+dedicatedPort)
	if err != nil {
		log.Printf("Error creating dedicated socket on port %s: %v", dedicatedPort, err)
		return err
	}
	defer dedicatedListener.Close()

	// Inform client of their dedicated port
	if _, err := conn.Write([]byte(dedicatedPort + "\n")); err != nil {
		log.Printf("Error sending port to client: %v", err)
		return err
	}

	// Accept connection from client on dedicated port
	clientConn, err := dedicatedListener.Accept()
	if err != nil {
		log.Printf("Error accepting client connection on dedicated port: %v", err)
		return err
	}

	log.Printf("Client %d connected on dedicated port %s", clientID, dedicatedPort)

	// Handle messages from this client in a goroutine
	go handleClientMessages(clientConn, dedicatedPort)

	return nil
}

// acceptConnectionsFromClients accepts and handles incoming client connections
func acceptConnectionsFromClients(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Check if the error is due to listener being closed (expected during shutdown)
			if strings.Contains(err.Error(), "use of closed network connection") {
				log.Println("Listener closed, shutting down accept loop")
				return nil
			}
			log.Printf("Error accepting connection: %v", err)
			return err
		}

		clientCounter++
		log.Printf("New connection from %s (client #%d)", conn.RemoteAddr(), clientCounter)

		go handleClientConnection(conn, clientCounter)
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found in parent directory")
	}

	// Get server port from environment or use default
	port = os.Getenv("SERVER_CONNECTIONS_PORT")
	if port == "" {
		log.Printf("Using default port %s", DefaultServerPort)
		port = DefaultServerPort
	}

	// Create main listener for client connections
	listener, err := createServerAcceptConnectionsSocket()
	if err != nil {
		log.Fatalf("Failed to create server socket: %v", err)
	}
	defer listener.Close()

	log.Printf("P2P Server %s running on port %s", ApplicationVersion, port)

	// Start accepting connections in background
	go func() {
		if err := acceptConnectionsFromClients(listener); err != nil {
			log.Printf("Accept loop error: %v", err)
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down server...")
}
