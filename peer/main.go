package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	common_helpers "P2P/common-helpers"
	"P2P/common-helpers/data"

	"github.com/joho/godotenv"
)

var (
	serverPort    string
	serverAddress string
	fileNames     []string
)

// loadConfig loads configuration from environment variables
func loadConfig() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: .env file not found in parent directory")
	}

	serverPort = os.Getenv("SERVER_CONNECTIONS_PORT")
	if serverPort == "" {
		log.Printf("Using default server port %s", DefaultServerPort)
		serverPort = DefaultServerPort
	}

	serverAddress = os.Getenv("SERVER_IP_ADDRESS")
	if serverAddress == "" {
		log.Println("Using default server address: localhost")
		serverAddress = "localhost"
	}
}

// connectToServer establishes initial connection and gets dedicated port
func connectToServer() (net.Conn, error) {
	// Initial connection to get dedicated port assignment
	initialConn, err := net.Dial("tcp", serverAddress+":"+serverPort)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	defer initialConn.Close()

	// Read dedicated port from server
	dedicatedPort, err := bufio.NewReader(initialConn).ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read dedicated port: %w", err)
	}

	dedicatedPort = strings.TrimSpace(dedicatedPort)
	log.Printf("Server assigned dedicated port: %s", dedicatedPort)

	// Connect to dedicated port
	dedicatedConn, err := net.Dial("tcp", serverAddress+":"+dedicatedPort)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to dedicated port: %w", err)
	}

	return dedicatedConn, nil
}

// loadRFCFiles loads available RFC files from the RFCs directory
func loadRFCFiles() error {
	entries, err := os.ReadDir("./RFCs")
	if err != nil {
		return fmt.Errorf("error reading RFC directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			fileNames = append(fileNames, entry.Name())
			log.Printf("Found RFC file: %s", entry.Name())
		}
	}

	return nil
}

// registerRFCs registers all available RFCs with the server
func registerRFCs(conn net.Conn, uploadPort string) error {
	reader := bufio.NewReader(conn)

	for _, filename := range fileNames {
		// Parse filename format: Number_title.txt
		parts := strings.Split(filename, "_")
		if len(parts) < 2 {
			log.Printf("Skipping invalid filename format: %s", filename)
			continue
		}

		rfcNumber := parts[0]
		rfcTitle := strings.TrimSuffix(parts[1], ".txt")

		addStruct := data.AddStruct{
			RFCNumber:                rfcNumber,
			RFCTitle:                 rfcTitle,
			ClientIP:                 conn.LocalAddr().String(),
			ClientUploadPort:         uploadPort,
			ClientApplicationVersion: ApplicationVersion,
		}

		serialized, err := SerializeAddStruct(addStruct)
		if err != nil {
			return fmt.Errorf("error serializing RFC %s: %w", rfcNumber, err)
		}

		message := append([]byte{byte(common_helpers.AddStructIndex)}, serialized...)
		message = append(message, '\n')

		if _, err := conn.Write(message); err != nil {
			return fmt.Errorf("error sending RFC %s: %w", rfcNumber, err)
		}

		// Read and consume the server response
		_, err = readServerResponse(reader, conn)
		if err != nil {
			log.Printf("Warning: Failed to read response for RFC %s: %v", rfcNumber, err)
		}

		log.Printf("Registered RFC %s: %s", rfcNumber, rfcTitle)
	}

	return nil
}

// startCommandLoop starts the interactive command loop
func startCommandLoop(conn net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	reader := bufio.NewReader(conn)
	for {
		fmt.Print("\nEnter command (ADD/LOOKUP/LIST/GET): ")

		if !scanner.Scan() {
			break
		}

		input := scanner.Text()


		if err := executeCommand(conn, input, reader); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Scanner error: %v", err)
	}
}

// sendErrorResponse sends an error response to the client
func sendErrorResponse(conn net.Conn, code int, phrase string) error {
	responseHeader := data.PeerResponseHeader{
		PeerApplicationVersion:    ApplicationVersion,
		Status:                    code,
		Phrase:                    phrase,
		CurrentDateandTime:        time.Now().Format(time.RFC3339),
		OS:                        runtime.GOOS,
		LastModifiedDateandTime:   "",
		ContentLength:             "0",
		ContentType:               "text/plain",
	}

	serialized, err := SerializePeerResponse(responseHeader,"")
	if err != nil {
		return fmt.Errorf("error serializing response: %w", err)
	}

	log.Printf("Error response created: %s", string(serialized))
	serialized = append(serialized, '\n')
	_, err = conn.Write(serialized)
	log.Printf("Error response sent to client: %s", string(serialized))
	return err
}

// sendSuccessResponse sends a success response with data to the client
func sendSuccessResponse(conn net.Conn, rfcNumber string) error {

	// Reload RFC files to include any newly added RFCs
	entries, err := os.ReadDir("./RFCs")
	if err != nil {
		log.Printf("Error reading RFC directory: %v", err)
		return sendErrorResponse(conn, 500, "Internal Server Error")
	}

	// Find the RFC file with the given number
	var rfcFilePath string
	for _, entry := range entries {
		if !entry.IsDir() {
			filename := entry.Name()
			if strings.HasPrefix(filename, rfcNumber+"_") {
				rfcFilePath = "./RFCs/" + filename
				break
			}
		}
	}

	fmt.Println("File path is: ", rfcFilePath)

	if rfcFilePath == "" {
		return sendErrorResponse(conn, 404, "RFC Not Found")
	}

	// Get file info for last modified time
	fileInfo, err := os.Stat(rfcFilePath)
	if err != nil {
		fmt.Println("Error getting file info: ", err)
		return sendErrorResponse(conn, 404, "RFC Not Found")
	}

	// Read the RFC file content
	responseData, err := os.ReadFile(rfcFilePath)
	if err != nil {
		return sendErrorResponse(conn, 500, "Internal Server Error")
	}

	responseHeader := data.PeerResponseHeader{
		PeerApplicationVersion:  ApplicationVersion,
		Status:                  200,
		Phrase:                  "OK",
		CurrentDateandTime:      time.Now().Format(time.RFC3339),
		OS:                      runtime.GOOS,
		LastModifiedDateandTime: fileInfo.ModTime().Format(time.RFC3339),
		ContentLength:           fmt.Sprintf("%d", len(responseData)),
		ContentType:             "text/plain",
	}

	serialized, err := SerializePeerResponse(responseHeader, string(responseData))
	if err != nil {
		return fmt.Errorf("error serializing response: %w", err)
	}

	serialized = append(serialized, '\n')
	_, err = conn.Write(serialized)
	return err
}

func handlePeerRequest(conn net.Conn) error {
	reader := bufio.NewReader(conn)

	peerRequest, err := reader.ReadBytes(byte('\n'))
	
	if err != nil {
		log.Printf("Error reading peer request: %v", err)
	}

	request, err := DeserializePeerRequest(peerRequest)
	if err != nil {
		return sendErrorResponse(conn, 400, "Bad Request")
	}

	if request.Version != ApplicationVersion {
		return sendErrorResponse(conn, 505, "P2P-CI Version Not Supported")
	}

	if request.PeerIP == conn.LocalAddr().String() {
		return sendErrorResponse(conn, 400, "Bad Request")
	}

	return sendSuccessResponse(conn, request.RFCNumber)
}

func main() {
	log.Println("P2P Client starting...")

	// Load configuration
	loadConfig()

	// Connect to server
	serverConn, err := connectToServer()


	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer serverConn.Close()

	log.Println("Successfully connected to server")

	// Load available RFC files
	if err := loadRFCFiles(); err != nil {
		log.Fatalf("Failed to load RFC files: %v", err)
	}

	// Get random port for upload server
	uploadPort, err := getRandomUploadPort()
	if err != nil {
		log.Fatalf("Failed to get upload port: %v", err)
	}
	log.Printf("Upload server will use port: %s", uploadPort)

	// Create upload listener
	uploadListener, err := net.Listen("tcp", ":"+uploadPort)
	if err != nil {
		log.Fatalf("Failed to create upload listener: %v", err)
	}
	defer uploadListener.Close()

	// Register all RFCs with server
	if err := registerRFCs(serverConn, uploadPort); err != nil {
		log.Fatalf("Failed to register RFCs: %v", err)
	}

	log.Println("All RFCs registered successfully")

	//This is the IP address of the host machine used to connect to the server
	hostIP := serverConn.LocalAddr().String()
	log.Printf("Host IP address: %s", hostIP)

	// Start command loop in goroutine
	go startCommandLoop(serverConn)

	//Set up shutdown signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	//On the main thread, we listen for requests on the upload port
	//The request is then desrialzed into the PeerRequest struct first
	//Then the request is handled by the handlePeerRequest function
	go func() {
		for {
			conn, err := uploadListener.Accept()
			if err != nil {
				// Check if error is due to listener being closed (expected during shutdown)
				if strings.Contains(err.Error(), "use of closed network connection") {
					log.Println("Upload listener closed, shutting down accept loop")
					return
				}
				log.Printf("Error accepting upload connection: %v", err)
				return
			}

			go func(c net.Conn) {
				defer c.Close()
				handlePeerRequest(c)
			}(conn)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Client Shutting down...")
	uploadListener.Close()
	serverConn.Close()
}
