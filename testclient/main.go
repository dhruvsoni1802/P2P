package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var port string
var serverAddress string

func main() {
	fmt.Println("Starting test client...")

	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: .env file not found in parent directory")
	}

	//Get the port from the environment file
	port = os.Getenv("SERVER_CONNECTIONS_PORT")
	if port == "" {
		log.Println("Port not found in environment file, Using default port 7734")
		port = "7734"
	}

	//Get the server address from the environment file
	serverAddress = os.Getenv("SERVER_ADDRESS")
	if serverAddress == "" {
		log.Println("Server address not found in environment file, Using default address localhost")
		serverAddress = "localhost"
	}

	//Connect to the server for registration
	conn, err := net.Dial("tcp", serverAddress+":"+port)

	if err != nil {
		log.Fatal("Error connecting to server: ", err)
	}
	defer conn.Close()

	//Blocking call to read the port from the server
	serverDedicatedport, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Fatal("Error reading port from server: ", err)
	}

	// Trim newline and whitespace from the received port
	serverDedicatedport = strings.TrimSpace(serverDedicatedport)

	fmt.Println("Connected to server and server is using port to communicate with me is ", serverDedicatedport)
	
	//Now we send a simple message to the server on the new port
	newconn, err := net.Dial("tcp", serverAddress+":"+serverDedicatedport)
	if err != nil {
		log.Fatal("Error connecting to server: ", err)
	}
	newconn.Write([]byte("Hello server from client on port " + serverDedicatedport + "\n"))
}