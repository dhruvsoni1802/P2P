// This file stores the application version and server configuration constants
package main

const (
	// ApplicationVersion is the P2P protocol version
	ApplicationVersion = "P2P-CI/1.0"

	// DefaultServerPort is the default port for accepting client connections
	DefaultServerPort = "7734"

	// HTTP status code equivalents for P2P protocol
	StatusOK                  = 200
	StatusBadRequest          = 400
	StatusNotFound            = 404
	StatusVersionNotSupported = 505
)
