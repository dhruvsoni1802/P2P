package main

import (
	"P2P/common-helpers/data"
	"encoding/json"
)

//Function to convert a byte array into an AddStruct
func DeserializeServerResponse(b []byte) (data.ServerResponse, error) {
	var serverResponse data.ServerResponse
	err := json.Unmarshal(b, &serverResponse)

	//If unmarshalling of the ServerResponse fails, this means that byte array got corrupted on network transmission
	if err != nil {
		return serverResponse, err
	}
	return serverResponse, nil
}