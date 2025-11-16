package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	common_helpers "P2P/common-helpers"
	"P2P/common-helpers/data"

	"github.com/joho/godotenv"
)

var port string
var clientCounter int = 0

// A global unordered map of Peer Info mapping hostname to upload port
var peerInfoMap = make(map[string]string)
var peerInfoMapMutex sync.RWMutex

//A global unordered map of RFC Indexes mapping from hostname to a list whose elements are a list of size 2 containing the RFC number and the RFC title
var rfcIndexMap = make(map[string][][]string)
var rfcIndexMapMutex sync.RWMutex

func createServerAcceptConnectionsSocket() (net.Listener, error) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, err
	}
	return listener, nil
}

func handleClientConnection(conn net.Conn, clientID int) error {
	//Now the server will create a new TCP socket on a port which is available from the global stack of free ports
	currentPort, err := common_helpers.GetFreePort()
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
		defer common_helpers.ReturnPort(currentPort)
		defer delete(peerInfoMap, connfromclient.RemoteAddr().String())
		defer delete(rfcIndexMap, connfromclient.RemoteAddr().String())

		reader := bufio.NewReader(connfromclient)
		for {
			message, err := reader.ReadBytes(byte('\n'))
			if err != nil {
				fmt.Println("Error reading message from client: ", err)
				return
			}

			//The first byte of the message is the index of the struct type
			structType := message[0]
			//Convert the byte to an int
			structTypeInt := int(structType)
			switch structTypeInt {
			case common_helpers.AddStructIndex:
				// Skip the first byte (struct type index) and the last byte (newline)
				jsonData := message[1 : len(message)-1]
				addStruct, err := DeserializeAddStruct(jsonData)

				//We will create a ServerResponseHeader and ServerResponseData structs
				serverResponseHeader := data.ServerResponseHeader{
					Response_Code: 200,
					Response_Phrase: "OK",
					Server_Application_Version: ApplicationVersion,
				}
				serverResponseData := data.ServerResponseData{
					RFC_Number: addStruct.RFC_Number,
					RFC_Title: addStruct.RFC_Title,
					Client_IP: addStruct.Client_IP,
					Client_Upload_Port: addStruct.Client_Upload_Port,
				}

				serverResponse := data.ServerResponse{
					Header: serverResponseHeader,
					Data: []data.ServerResponseData{},
				}

				if err != nil {
					fmt.Println("Error deserializing AddStruct: ", err)
					serverResponse.Header.Response_Code = 400
					serverResponseHeader.Response_Phrase = "Bad Request"

					//We need to serialize the serverResponse struct into a byte array and then send it back to the client
					serializedServerResponse, err := SerializeServerResponse(serverResponse)
					if err != nil {
						log.Println("Error serializing ServerResponse: ", err)
					}

					serializedServerResponse = append(serializedServerResponse, '\n')
					connfromclient.Write(serializedServerResponse)
					return
				}


				fmt.Println("AddStruct: RFC_Number: ", addStruct.RFC_Number, " RFC_Title: ", addStruct.RFC_Title, " Client_IP: ", addStruct.Client_IP, " Client_Upload_Port: ", addStruct.Client_Upload_Port, " Client_Application_Version: ", addStruct.Client_Application_Version)

				//We check if the client application version is same as the server application version
				//If not, we send a response back to the client with the response code 505 and the response phrase "P2P-CI Version Not Supported"
				if addStruct.Client_Application_Version != ApplicationVersion {
					log.Println("Application version mismatch: ", addStruct.Client_Application_Version, " != ", ApplicationVersion)
					serverResponse.Header.Response_Code = 505
					serverResponseHeader.Response_Phrase = "P2P-CI Version Not Supported"

					//We need to serialize the serverResponse struct into a byte array and then send it back to the client
					serializedServerResponse, err := SerializeServerResponse(serverResponse)
					if err != nil {
						log.Println("Error serializing ServerResponse: ", err)
					}

					serializedServerResponse = append(serializedServerResponse, '\n')
					connfromclient.Write(serializedServerResponse)
					return
				}

				//If the rfc index map doesn't have the client IP, create a new list first
				if _, ok := rfcIndexMap[addStruct.Client_IP]; !ok {
					//Add a RW Mutex for adding information to the rfcIndexMap
					rfcIndexMapMutex.Lock()
					defer rfcIndexMapMutex.Unlock()
					rfcIndexMap[addStruct.Client_IP] = make([][]string, 0)
				}

				// Add the RFC number and title to the list
				// If the RFC number and title are already in the list, no need to add them again
				alreadyExists := false

				//TODO: Optimize this to use a hash map instead of a linear search
				for _, rfcInfo := range rfcIndexMap[addStruct.Client_IP] {
					if len(rfcInfo) == 2 && rfcInfo[0] == addStruct.RFC_Number && rfcInfo[1] == addStruct.RFC_Title {
						fmt.Println("RFC number and title already in list, no need to add them again")
						alreadyExists = true
						break
					}
				}
				if alreadyExists {
					continue
				}
				hostname := strings.Split(addStruct.Client_IP, ":")[0]
				rfcIndexMap[hostname] = append(rfcIndexMap[hostname], []string{addStruct.RFC_Number, addStruct.RFC_Title})
				fmt.Println("Client IP: ", hostname, " RFC number and title added to list: ", addStruct.RFC_Number, " ", addStruct.RFC_Title)

				//If the client IP is already in the map, no need to add it again
				if _, ok := peerInfoMap[addStruct.Client_IP]; ok {
					fmt.Println("Client IP already in map, no need to add it again")
					continue
				}

				//Add a RW Mutex for adding information to the peerInfoMap
				peerInfoMapMutex.Lock()
				defer peerInfoMapMutex.Unlock()
				peerInfoMap[hostname] = addStruct.Client_Upload_Port
				fmt.Println("Client IP added to map: ", hostname, " ", addStruct.Client_Upload_Port)

				serverResponse.Data = append(serverResponse.Data, serverResponseData)
				serializedServerResponse, err := SerializeServerResponse(serverResponse)
				if err != nil {
						log.Println("Error serializing ServerResponse: ", err)
				}

				serializedServerResponse = append(serializedServerResponse, '\n')
				connfromclient.Write(serializedServerResponse)

			case common_helpers.LookupStructIndex:
				// Skip the first byte (struct type index) and the last byte (newline)
				jsonData := message[1 : len(message)-1]
				lookUpStruct, err := DeserializeLookUpStruct(jsonData)
				if err != nil {
					log.Println("Error deserializing LookUpStruct: ", err)
					return
				}


				fmt.Println("LookUpStruct: RFC_Number: ", lookUpStruct.RFC_Number, " RFC_Title: ", lookUpStruct.RFC_Title, " Client_IP: ", lookUpStruct.Client_IP, " Client_Upload_Port: ", lookUpStruct.Client_Upload_Port, " Client_Application_Version: ", lookUpStruct.Client_Application_Version)
				
			case common_helpers.ListStructIndex:
				// Skip the first byte (struct type index) and the last byte (newline)
				jsonData := message[1 : len(message)-1]
				listStruct, err := DeserializeListStruct(jsonData)
				if err != nil {
					log.Println("Error deserializing ListStruct: ", err)
					return
				}
				fmt.Println("ListStruct: Client_IP: ", listStruct.Client_IP, " Client_Upload_Port: ", listStruct.Client_Upload_Port, " Client_Application_Version: ", listStruct.Client_Application_Version)
			}
			
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

	//The main thread is the receiver of the interrupt signal from the user and blocks here
	<-sigChan

	fmt.Println("Server is shutting down...")
	listener.Close()
}