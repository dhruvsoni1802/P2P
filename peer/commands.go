package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	common_helpers "P2P/common-helpers"
	"P2P/common-helpers/data"
)

// CommandType represents the type of P2P command
type CommandType string

const (
	CommandAdd    CommandType = "ADD"
	CommandLookup CommandType = "LOOKUP"
	CommandList   CommandType = "LIST"
	CommandGet    CommandType = "GET"
)

// Command represents a parsed user command
type Command struct {
	Type    CommandType
	RFC     string
	Version string
	DataSection map[string]string
}

// parseCommand parses a raw command string into a Command struct
func parseCommand(input string) (*Command, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty command")
	}

	parts := strings.Fields(input)
	if len(parts) < 3 {
		return nil, fmt.Errorf("insufficient arguments")
	}

	method := strings.ToUpper(parts[0])
	if method == "GET" {
		return &Command{
			Type: CommandGet,
			RFC: "",
			Version: "",
			DataSection:nil,
		}, nil
	}
	if method != "ADD" && method != "LOOKUP" && method != "LIST" {
		return nil, fmt.Errorf("invalid method: must be ADD, LOOKUP, or LIST")
	}

	rfcString := parts[1]
	rfcNumber := parts[2]

	if method == "LIST" && rfcString != "ALL" {
		return nil, fmt.Errorf("LIST requires ALL parameter")
	}
	if method != "LIST" && rfcString != "RFC"  {
		return nil, fmt.Errorf("ADD and LOOKUP require RFC parameter")
	}

	if method != "LIST" && !isNumeric(rfcNumber) {
		return nil, fmt.Errorf("ADD and LOOKUP require numeric RFC number")
	}

	var version string
	if method == "LIST" {
		version = parts[2]
	} else {
		version = parts[3]
	}

	// Parse data section of the command
	dataSection := make(map[string]string)
	for i := 3; i < len(parts); i++ {
		if strings.Contains(parts[i], ":") {
			kv := strings.SplitN(parts[i], ":", 2)
			if len(kv) == 2 {
				dataSection[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	// Validate required headers
	if _, ok := dataSection["Host"]; !ok {
		return nil, fmt.Errorf("missing Host header")
	}
	if _, ok := dataSection["Port"]; !ok {
		return nil, fmt.Errorf("missing Port header")
	}
	if _, ok := dataSection["Title"]; !ok && method != "LIST" {
		return nil, fmt.Errorf("missing Title header")
	}

	return &Command{
		Type:    CommandType(method),
		RFC:     rfcNumber,
		Version: version,
		DataSection: dataSection,
	}, nil
}

func sendGetCommand(input string) (data.PeerResponseHeader, string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return data.PeerResponseHeader{}, "", fmt.Errorf("empty command")
	}

	parts := strings.Fields(input)
	if len(parts) < 4 {
		return data.PeerResponseHeader{}, "", fmt.Errorf("insufficient arguments")
	}

	method := strings.ToUpper(parts[0])
	if method != "GET" {
		return data.PeerResponseHeader{}, "", fmt.Errorf("invalid method: must be GET")
	}

	rfcString := parts[1]
	if rfcString != "RFC" {
		return data.PeerResponseHeader{}, "", fmt.Errorf("GET requires RFC parameter")
	}

	rfcNumber := parts[2]
	if !isNumeric(rfcNumber) {
		return data.PeerResponseHeader{}, "", fmt.Errorf("RFC number must be numeric")
	}

	version := parts[3]

	// Parse data section of the command
	dataSection := make(map[string]string)
	for i := 4; i < len(parts); i++ {
		if strings.Contains(parts[i], ":") {
			kv := strings.SplitN(parts[i], ":", 2)
			if len(kv) == 2 {
				dataSection[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	// Validate required headers
	if _, ok := dataSection["Host"]; !ok {
		return data.PeerResponseHeader{}, "", fmt.Errorf("missing Host header")
	}
	if _, ok := dataSection["OS"]; !ok {
		return data.PeerResponseHeader{}, "", fmt.Errorf("missing OS header")
	}

	//Now we create a new TCP socket to make the GET request to the other peer
	hostIP := strings.Split(dataSection["Host"], ":")[0]
	hostPort := strings.Split(dataSection["Host"], ":")[1]

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", hostIP, hostPort))
	if err != nil {
		return data.PeerResponseHeader{}, "", fmt.Errorf("error connecting to peer: %w", err)
	}

	//Now we first figure out on which port we just created the TCP socket
	localAddr := conn.LocalAddr()

	//Now we send the GET request to the other peer
	request := data.PeerRequest{
		RFCNumber: rfcNumber,
		Version: version,
		PeerIP: localAddr.String(),
		PeerOS: dataSection["OS"],
	}

	//Now we serialize the request
	serializedRequest, err := SerializePeerRequest(request)
	if err != nil {
		return data.PeerResponseHeader{}, "", fmt.Errorf("error serializing peer request: %w", err)
	}

	//Now we send the request to the other peer
	message := append(serializedRequest, '\n')
	if _, err := conn.Write(message); err != nil {
		return data.PeerResponseHeader{}, "", fmt.Errorf("error sending GET request: %w", err)
	}

	fmt.Println("GET request sent successfully")

	//Now we wait for the peer response which is the the peer response struct
	reader := bufio.NewReader(conn)
	peerResponseHeader, peerResponseData, err := readPeerResponse(reader, conn)
	if err != nil {
		return data.PeerResponseHeader{}, "",	 fmt.Errorf("error reading peer response: %w", err)
	}

	return peerResponseHeader, peerResponseData, nil
}

//Format the server response converting the struct to a string
func formatServerResponse(serverResponse data.ServerResponse) string {
	var result strings.Builder

	// First line: version <sp> status code <sp> phrase <cr> <lf>
	result.WriteString(fmt.Sprintf("%s %d %s\r\n",
		serverResponse.Header.ServerApplicationVersion,
		serverResponse.Header.ResponseCode,
		serverResponse.Header.ResponsePhrase))

	// For each RFC in the data array
	for _, rfcData := range serverResponse.Data {
		result.WriteString(fmt.Sprintf("%s %s %s %s\r\n",
			rfcData.RFCNumber,
			rfcData.RFCTitle,
			rfcData.ClientIP,
			rfcData.ClientUploadPort))
	}

	return result.String()
}

//Format the peer response header converting the struct to a string
func formatPeerResponse(peerResponseHeader data.PeerResponseHeader, peerResponseData string) string {
	var result strings.Builder

	// First line: version <sp> status code <sp> phrase
	result.WriteString(fmt.Sprintf("%s %d %s\r\n",
		peerResponseHeader.PeerApplicationVersion,
		peerResponseHeader.Status,
		peerResponseHeader.Phrase))

	// Date header
	result.WriteString(fmt.Sprintf("Date: %s\r\n", peerResponseHeader.CurrentDateandTime))

	// OS header
	result.WriteString(fmt.Sprintf("OS: %s\r\n", peerResponseHeader.OS))

	// Last-Modified header
	result.WriteString(fmt.Sprintf("Last-Modified: %s\r\n", peerResponseHeader.LastModifiedDateandTime))

	// Content-Length header
	result.WriteString(fmt.Sprintf("Content-Length: %s\r\n", peerResponseHeader.ContentLength))

	// Content-Type header
	result.WriteString(fmt.Sprintf("Content-Type: %s\r\n", peerResponseHeader.ContentType))

	// Empty line before data
	result.WriteString("\r\n")

	// Data
	result.WriteString(peerResponseData)

	return result.String()
}

func readPeerResponse(reader *bufio.Reader, conn net.Conn) (data.PeerResponseHeader, string, error) {
	fmt.Println("Reading peer response")
	conn.SetReadDeadline(time.Now().Add(PeerResponseTimeout))
	 
	peerResponseRaw, err := reader.ReadBytes(byte('\n'))
	fmt.Println("Peer response read successfully")
	if err != nil {
		return data.PeerResponseHeader{}, "", fmt.Errorf("error reading peer response: %w", err)
	}
	conn.SetReadDeadline(time.Time{})
	peerResponseHeader, peerResponseData, err := DeserializePeerResponseData(peerResponseRaw)
	fmt.Println("Peer response deserialized successfully")
	if err != nil {
		return data.PeerResponseHeader{}, "", fmt.Errorf("error deserializing peer response: %w", err)
	}

	return peerResponseHeader, peerResponseData, nil
}

// readServerResponse reads a server response from the connection
func readServerResponse(reader *bufio.Reader, conn net.Conn) (data.ServerResponse, error) {
	conn.SetReadDeadline(time.Now().Add(ServerResponseTimeout))
	
	serverResponse, err := reader.ReadBytes(byte('\n'))
	
	if err != nil {
			return data.ServerResponse{}, fmt.Errorf("error reading server response: %w", err)
	}
	conn.SetReadDeadline(time.Time{})
	serverResponseData, err := DeserializeServerResponse(serverResponse)
	if err != nil {
			return data.ServerResponse{}, fmt.Errorf("error deserializing server response: %w", err)
	}

	return serverResponseData, nil
}

// sendAddRequest sends an ADD request to the server
func sendAddRequest(conn net.Conn, cmd *Command, reader *bufio.Reader) error {
	addStruct := data.AddStruct{
		RFCNumber:                cmd.RFC,
		RFCTitle:                 cmd.DataSection["Title"],
		ClientIP:                 cmd.DataSection["Host"],
		ClientUploadPort:         cmd.DataSection["Port"],
		ClientApplicationVersion: cmd.Version,
	}

	serialized, err := SerializeAddStruct(addStruct)
	if err != nil {
		return fmt.Errorf("error serializing AddStruct: %w", err)
	}

	message := append([]byte{byte(common_helpers.AddStructIndex)}, serialized...)
	message = append(message, '\n')

	if _, err := conn.Write(message); err != nil {
		return fmt.Errorf("error sending ADD request: %w", err)
	}

	fmt.Println("ADD request sent successfully")

	//Now we wait for the server response
	serverResponse, err := readServerResponse(reader, conn)
	serverResponseString := formatServerResponse(serverResponse)
	fmt.Printf("Server response:\n%s", serverResponseString)
	if err != nil { 
		return fmt.Errorf("error reading server response: %w", err)
	}

	switch serverResponse.Header.ResponseCode {
	case StatusOK:
		fmt.Println("RFC added successfully")
	case StatusBadRequest:
		fmt.Println("Error: Bad Request")
	case StatusVersionNotSupported:
		fmt.Println("Error: P2P-CI Version Not Supported")
	default:
		fmt.Println("Error: Unknown server response code")
	}
	return nil
}

// sendLookupRequest sends a LOOKUP request to the server
func sendLookupRequest(conn net.Conn, cmd *Command, reader *bufio.Reader) error {
	lookupStruct := data.LookUpStruct{
		RFCNumber:                cmd.RFC,
		RFCTitle:                 cmd.DataSection["Title"],
		ClientIP:                 cmd.DataSection["Host"],
		ClientUploadPort:         cmd.DataSection["Port"],
		ClientApplicationVersion: cmd.Version,
	}

	serialized, err := SerializeLookUpStruct(lookupStruct)
	if err != nil {
		return fmt.Errorf("error serializing LookUpStruct: %w", err)
	}

	message := append([]byte{byte(common_helpers.LookupStructIndex)}, serialized...)
	message = append(message, '\n')

	if _, err := conn.Write(message); err != nil {
		return fmt.Errorf("error sending LOOKUP request: %w", err)
	}

	fmt.Println("LOOKUP request sent successfully")

	//Now we wait for the server response
	serverResponse, err := readServerResponse(reader, conn)
	serverResponseString := formatServerResponse(serverResponse)
	if err != nil {
		return fmt.Errorf("error reading server response: %w", err)
	}

	fmt.Printf("Server response:\n%s", serverResponseString)

	//TODO: Remove later
	switch serverResponse.Header.ResponseCode {
	case StatusOK:
		fmt.Println("RFC lookup response received successfully")
	case StatusBadRequest:
		fmt.Println("Error: Bad Request")
	case StatusNotFound:
		fmt.Println("Error: Not Found")
	case StatusVersionNotSupported:
		fmt.Println("Error: P2P-CI Version Not Supported")
	default:
		fmt.Println("Error: Unknown server response code")
	}
	return nil
}

// sendListRequest sends a LIST request to the server
func sendListRequest(conn net.Conn, cmd *Command, reader *bufio.Reader) error {
	listStruct := data.ListStruct{
		ClientIP:                 cmd.DataSection["Host"],
		ClientUploadPort:         cmd.DataSection["Port"],
		ClientApplicationVersion: cmd.Version,
	}

	serialized, err := SerializeListStruct(listStruct)
	if err != nil {
		return fmt.Errorf("error serializing ListStruct: %w", err)
	}

	message := append([]byte{byte(common_helpers.ListStructIndex)}, serialized...)
	message = append(message, '\n')

	if _, err := conn.Write(message); err != nil {
		return fmt.Errorf("error sending LIST request: %w", err)
	}

	fmt.Println("LIST request sent successfully")

	//Now we wait for server response
	//Now we wait for the server response
	serverResponse, err := readServerResponse(reader, conn)
	if err != nil {
		return fmt.Errorf("error reading server response: %w", err)
	}

	serverResponseString := formatServerResponse(serverResponse)
	fmt.Printf("Server response:\n%s", serverResponseString)

	switch serverResponse.Header.ResponseCode {
	case StatusOK:
		fmt.Println("RFC list response received successfully")
	case StatusBadRequest:
		fmt.Println("Error: Bad Request")
	case StatusVersionNotSupported:
		fmt.Println("Error: P2P-CI Version Not Supported")
	default:
		fmt.Println("Error: Unknown server response code")
	}
	return nil

}

// executeCommand parses and executes a command
func executeCommand(conn net.Conn, input string, reader *bufio.Reader) error {
	cmd, err := parseCommand(input)
	if err != nil {
		return err
	}

	switch cmd.Type {
	case CommandAdd:
		return sendAddRequest(conn, cmd, reader)
	case CommandLookup:
		return sendLookupRequest(conn, cmd, reader)
	case CommandList:
		return sendListRequest(conn, cmd, reader)
	case CommandGet:
		var wg sync.WaitGroup
		var peerResponseHeader data.PeerResponseHeader
		var peerResponseData string
		var getErr error

		wg.Add(1)
		go func() {
			defer wg.Done()
			peerResponseHeader, peerResponseData, getErr = sendGetCommand(input)
			if getErr != nil {
				fmt.Printf("Error sending GET request: %v\n", getErr)
				return
			}
		}()

		wg.Wait()

		if getErr != nil {
			return getErr
		}

		// Format and display the peer response
		formattedResponse := formatPeerResponse(peerResponseHeader, peerResponseData)
		fmt.Printf("%s\n", formattedResponse)

		return nil
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}
