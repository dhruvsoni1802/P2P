package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
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

	version := parts[3]

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
	if _, ok := dataSection["Title"]; !ok {
		return nil, fmt.Errorf("missing Title header")
	}

	return &Command{
		Type:    CommandType(method),
		RFC:     rfcNumber,
		Version: version,
		DataSection: dataSection,
	}, nil
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
		ClientApplicationVersion: ApplicationVersion,
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
	return nil
}

// sendListRequest sends a LIST request to the server
func sendListRequest(conn net.Conn, cmd *Command, reader *bufio.Reader) error {
	listStruct := data.ListStruct{
		ClientIP:                 cmd.DataSection["Host"],
		ClientUploadPort:         cmd.DataSection["Port"],
		ClientApplicationVersion: ApplicationVersion,
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
	default:
		return fmt.Errorf("unknown command type: %s", cmd.Type)
	}
}
