package main

import (
	"P2P/common-helpers/data"
	"encoding/json"
)

//Function to convert ServerResponse into a byte array
func SerializeServerResponse(serverResponse data.ServerResponse) ([]byte, error) {
	json, err := json.Marshal(serverResponse)
	if err != nil {
		return nil, err
	}
	return json, nil
}