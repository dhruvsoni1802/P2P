package main

import (
	common_helpers "P2P/common-helpers"
	"math/rand"
	"strconv"
)

//Function to get a random Port which will be used for uploading files when requested from other server
func getRandomUploadPort() (string, error) {
	port := strconv.Itoa(rand.Intn(3001) + 4000)

	// Check if the port is available, otherwise keep trying
	for !common_helpers.IsPortAvailable(port) {
		port = strconv.Itoa(rand.Intn(3001) + 4000)
	}
	return port, nil
}