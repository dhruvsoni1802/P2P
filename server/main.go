package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
)

var port string
var clientCounter int = 0

func createServerAcceptConnectionsSocket() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func handleClientConnection(conn net.Conn, clientID int) error {
	//Now the server will create a new TCP socket on a port which is available from the global stack of free ports
	currentPort, err := getFreePort()
	if err != nil {
		log.Printf("Error getting free port for client %d: %v", clientID, err)
		return err
	}

	fmt.Println("Client ", clientID, " is using port ", currentPort)

	//Create a new TCP socket on this port
	serverSocketlistenerforClient, err := net.Listen("tcp", ":"+currentPort)
	if err != nil {
		fmt.Println("Error creating server socket on port ", currentPort, ": ", err)
		return err
	}

	fmt.Println("Server socket created on port ", currentPort)
	defer serverSocketlistenerforClient.Close()

	// Now we need to tell the client about the port the server is using for communication with it
	conn.Write([]byte(currentPort + "\n"))

  //Now we need to accept connection from the client on the new socket
	connfromclient, err := serverSocketlistenerforClient.Accept()
		if err != nil {
			log.Println("Error accepting connection from client: ", err)
			return err
		}

	//Now we need to listen for messages from the client on the new socket
	go func() {
		defer connfromclient.Close()
		reader := bufio.NewReader(connfromclient)
		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					log.Printf("Client %d disconnected from port %s", clientID, currentPort)
					returnPort(currentPort)
				} else {
					log.Printf("Error reading message from client %d: %v", clientID, err)
				}
				return
			}
			fmt.Println("Message from client ", clientID, " on port ", currentPort, " is: ", strings.TrimSpace(message))
		}
	}()	
	return nil
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
		clientCounter++

		go handleClientConnection(conn,clientCounter)
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