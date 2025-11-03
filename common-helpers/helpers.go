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

// A global stack which stores all the free ports
var (
	freePortsStack []string
	portStackMu    sync.Mutex
)

func init() {
	for i := 4000; i < 7000; i++ {
		freePortsStack = append(freePortsStack, strconv.Itoa(i))
	}
}

// Check if a port is available by attempting to listen on it
func IsPortAvailable(port string) bool {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}


// Get a free port from the stack, checking availability
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

// Return a port back to the pool when client disconnects
func ReturnPort(port string) {
	portStackMu.Lock()
	defer portStackMu.Unlock()

	// Add back to the stack
	freePortsStack = append(freePortsStack, port)
	log.Printf("Returned port %s to pool (Available: %d)", port, len(freePortsStack))
}

//Helper function to read the file names from the client 
func ReadFileNamesFromClient(conn net.Conn) ([]string, error) {
	reader := bufio.NewReader(conn)
	message, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	return strings.Split(message, ","), nil
}