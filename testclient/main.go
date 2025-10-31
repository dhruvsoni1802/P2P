package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/joho/godotenv"
)

var port string

func main() {
	fmt.Println("Starting test client...")

	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Warning: .env file not found in parent directory")
	}

	port = os.Getenv("SERVER_CONNECTIONS_PORT")
	if port == "" {
		log.Println("Port not found in environment file, Using default port 7734")
		port = "7734"
	}

	//Connect to the server
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal("Error connecting to server: ", err)
	}

	fmt.Println("Connected to server")
	
	//Close the connection
	conn.Close()
}