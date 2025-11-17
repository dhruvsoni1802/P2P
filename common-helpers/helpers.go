package common_helpers

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

const (
	// Port range constants for server-client communication
	MinPortRange = 4000
	MaxPortRange = 7000

	// Protocol struct type indices for message identification
	AddStructIndex    = 1
	ListStructIndex   = 2
	LookupStructIndex = 3
)

// A global stack which stores all the free ports
var (
	freePortsStack []string
	portStackMu    sync.Mutex
)

func init() {
	for i := MinPortRange; i < MaxPortRange; i++ {
		freePortsStack = append(freePortsStack, strconv.Itoa(i))
	}
}

// IsPortAvailable checks if a port is available by attempting to listen on it.
// Returns true if the port is available, false otherwise.
func IsPortAvailable(port string) bool {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// GetFreePort retrieves an available port from the pool for server-client communication.
// It checks each port for availability and returns the first available one.
// Returns an error if no ports are available in the pool.
func GetFreePort() (string, error) {
	portStackMu.Lock()
	defer portStackMu.Unlock()

	// Keep trying ports until we find an available one
	for len(freePortsStack) > 0 {
		port := freePortsStack[len(freePortsStack)-1]
		freePortsStack = freePortsStack[:len(freePortsStack)-1]

		// Check if port is actually available
		if IsPortAvailable(port) {
			return port, nil
		}
	}

	// No ports available
	return "", fmt.Errorf("no free ports available in pool")
}

// ReturnPort returns a port back to the pool when a client disconnects.
// This makes the port available for reuse by other clients.
func ReturnPort(port string) {
	portStackMu.Lock()
	defer portStackMu.Unlock()

	// Add back to the stack
	freePortsStack = append(freePortsStack, port)
	log.Printf("Returned port %s to pool (Available: %d)", port, len(freePortsStack))
}

// ReadFileNamesFromClient reads comma-separated file names from a client connection.
// Returns a slice of file names or an error if the read fails.
func ReadFileNamesFromClient(conn net.Conn) ([]string, error) {
	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	return strings.Split(message, ","), nil
}

// MessageReader wraps a buffered reader for reading protocol messages
type MessageReader struct {
	reader *bufio.Reader
}

// NewMessageReader creates a new MessageReader for the given connection
func NewMessageReader(conn net.Conn) *MessageReader {
	return &MessageReader{
		reader: bufio.NewReader(conn),
	}
}

// ReadMessage reads a single newline-terminated message from the connection
func (mr *MessageReader) ReadMessage() ([]byte, error) {
	return mr.reader.ReadBytes('\n')
}
