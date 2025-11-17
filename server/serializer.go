package main

import (
	"P2P/common-helpers/data"
	"encoding/json"
)

// SerializeServerResponse converts ServerResponse into a JSON byte array
func SerializeServerResponse(serverResponse data.ServerResponse) ([]byte, error) {
	jsonData, err := json.Marshal(serverResponse)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}
