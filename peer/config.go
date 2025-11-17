// This file stores the application version and client configuration constants
package main

import "time"

const (
	// ApplicationVersion is the P2P protocol version
	ApplicationVersion = "P2P-CI/1.0"

	// DefaultServerPort is the default port for connecting to server
	DefaultServerPort = "7734"

	// ServerResponseTimeout is the timeout for waiting for server responses
	ServerResponseTimeout = 5 * time.Second

	// HTTP status code equivalents for P2P protocol
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatusVersionNotSupported = 505
)
