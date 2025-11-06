package main

import (
	common_helpers "P2P/common-helpers"
	"P2P/common-helpers/data"
	"bufio"
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
var serverAddress string
var fileNames[] string

func main() {
	fmt.Println("Starting test client...")

	//Keep the connection alive until user interrupt signal
	// We need to block the main go routine to keep the client running
	//Main thread will exit on user interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

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
	serverAddress = os.Getenv("SERVER_IP_ADDRESS")
	if serverAddress == "" {
		log.Println("Server address not found in environment file, Using default address localhost")
		serverAddress = "localhost"
	}

	//Connect to the server for registration on 7734
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
	defer newconn.Close()

	//Retrieve the available filenames from the RFC directory
	filenames, err := os.ReadDir("./RFCs")
	if err != nil {
		log.Fatal("Error reading RFC directory: ", err)
	}

	for _, filename := range filenames {
		fmt.Println("Available filename: ", filename.Name())
		fileNames = append(fileNames, filename.Name())
	}

	//We take the random port for uploading files here
	randomPort, err := getRandomUploadPort()
	if err != nil {
		log.Fatal("Error getting random port: ", err)
	}

	fmt.Println("Random port for uploading files is: ", randomPort)

	//Now we create a new TCP socket on the random port
	randomPortSocket, err := net.Listen("tcp", ":"+randomPort)
	if err != nil {
		log.Fatal("Error creating new TCP socket on random port: ", err)
	}
	defer randomPortSocket.Close()

	//We will iterate over the file names
	for _, filename := range fileNames {

		//Now we create a new AddStruct for the client
		//Split the filename which is of format Number_title.txt
		//Trim the title to remove the .txt extension
		filenameParts := strings.Split(filename, "_")
		RFC_Number := filenameParts[0]
		RFC_Title := filenameParts[1]
		RFC_Title = strings.TrimSuffix(RFC_Title, ".txt")

		fmt.Println("RFC_Number: ", RFC_Number, " RFC_Title: ", RFC_Title)


		addStruct := data.AddStruct{
			RFC_Number: RFC_Number,
			RFC_Title: RFC_Title,
			Client_IP: newconn.RemoteAddr().String(),
			Client_Upload_Port: randomPort,
			Client_Application_Version: ApplicationVersion,
		}

		serializedAddStruct, err := SerializeAddStruct(addStruct)
		if err != nil {
			log.Fatal("Error serializing AddStruct: ", err)
		}

		//Now we send the serialized AddStruct to the server
		//Add the index of the struct type (AddStructIndex) at the start of the serializedAddStruct
		serializedAddStruct = append([]byte{byte(common_helpers.AddStructIndex)}, serializedAddStruct...)
		serializedAddStruct = append(serializedAddStruct, '\n')
		newconn.Write(serializedAddStruct)
	}

	//A while true loop, which waits for a user input command
	go func() {
		for {
			fmt.Print("Enter command: ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			input := strings.TrimSpace(scanner.Text())
			
			if input == "" {
				fmt.Println("Error: empty command")
				continue
			}
			
			parts := strings.Fields(input)
			
			// Validate format: METHOD RFC/ALL VERSION Host:value Port:value Title:value
			if len(parts) < 3 {
				fmt.Println("Error: insufficient arguments")
				continue
			}
			
			method := strings.ToUpper(parts[0])
			if method != "ADD" && method != "LOOKUP" && method != "LIST" {
				fmt.Println("Error: invalid method")
				continue
			}
			
			rfc := parts[1]
			if method == "LIST" && rfc != "ALL" {
				fmt.Println("Error: LIST requires ALL parameter")
				continue
			}
			if method != "LIST" && !strings.HasPrefix(rfc, "RFC") {
				fmt.Println("Error: invalid RFC format")
				continue
			}
			
			// version := parts[2]
			// if version != "P2P-CI/1.0" {
			// 	fmt.Println("Error: invalid version")
			// 	continue
			// }
			
			// Parse headers
			headers := make(map[string]string)
			for i := 3; i < len(parts); i++ {
				if strings.Contains(parts[i], ":") {
					kv := strings.SplitN(parts[i], ":", 2)
					if len(kv) == 2 {
						headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
					}
				}
			}
			
			// Validate required headers
			if _, ok := headers["Host"]; !ok {
				fmt.Println("Error: missing Host header")
				continue
			}
			if _, ok := headers["Port"]; !ok {
				fmt.Println("Error: missing Port header")
				continue
			}
			if _, ok := headers["Title"]; !ok {
				fmt.Println("Error: missing Title header")
				continue
			}
			
			fmt.Println("Valid input from the user")

			if method == "ADD" {
				//Create a new AddStruct
				addStruct := data.AddStruct{
					RFC_Number: parts[2],
					RFC_Title: headers["Title"],
					Client_IP: headers["Host"],
					Client_Upload_Port: headers["Port"],
					Client_Application_Version: ApplicationVersion,
				}

				serializedAddStruct, err := SerializeAddStruct(addStruct)
				if err != nil {
					log.Fatal("Error serializing AddStruct: ", err)
				}

				serializedAddStruct = append([]byte{byte(common_helpers.AddStructIndex)}, serializedAddStruct...)
				serializedAddStruct = append(serializedAddStruct, '\n')
				newconn.Write(serializedAddStruct)
				fmt.Println("AddStruct sent to server")
			}
			if method == "LOOKUP" {
				//Create a new LookUpStruct
				lookUpStruct := data.LookUpStruct{
					RFC_Number: parts[2],
					RFC_Title: headers["Title"],
					Client_IP: headers["Host"],
					Client_Upload_Port: headers["Port"],
					Client_Application_Version: ApplicationVersion,
				}

				serializedLookUpStruct, err := SerializeLookUpStruct(lookUpStruct)
				if err != nil {
					log.Fatal("Error serializing LookUpStruct: ", err)
				}

				serializedLookUpStruct = append([]byte{byte(common_helpers.LookupStructIndex)}, serializedLookUpStruct...)
				serializedLookUpStruct = append(serializedLookUpStruct, '\n')
				newconn.Write(serializedLookUpStruct)
				fmt.Println("LookUpStruct sent to server")
			}
			if method == "LIST" {
				//Create a new ListStruct
				listStruct := data.ListStruct{
					Client_IP: headers["Host"],
					Client_Upload_Port: headers["Port"],
					Client_Application_Version: ApplicationVersion,
				}

				serializedListStruct, err := SerializeListStruct(listStruct)
				if err != nil {
					log.Fatal("Error serializing ListStruct: ", err)
				}

				serializedListStruct = append([]byte{byte(common_helpers.ListStructIndex)}, serializedListStruct...)
				serializedListStruct = append(serializedListStruct, '\n')
				newconn.Write(serializedListStruct)
				fmt.Println("ListStruct sent to server")
			}
		}
	}()

	<-sigChan
	fmt.Println("Client is shutting down...")
	os.Exit(0)
}