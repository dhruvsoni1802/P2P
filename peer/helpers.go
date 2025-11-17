package main

import (
	common_helpers "P2P/common-helpers"
	"math/rand"
	"strconv"
	"unicode"
)

// Function to get a random Port which will be used for uploading files when requested from other server
func getRandomUploadPort() (string, error) {
	port := strconv.Itoa(rand.Intn(3001) + 4000)

	// Check if the port is available, otherwise keep trying
	for !common_helpers.IsPortAvailable(port) {
		port = strconv.Itoa(rand.Intn(3001) + 4000)
	}
	return port, nil
}

// isNumeric checks if a string contains only numeric characters
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
