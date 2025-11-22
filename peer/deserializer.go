package main

import (
	"P2P/common-helpers/data"
	"encoding/json"
)

// DeserializeServerResponse converts a JSON byte array into a ServerResponse struct
func DeserializeServerResponse(b []byte) (data.ServerResponse, error) {
	var serverResponse data.ServerResponse
	err := json.Unmarshal(b, &serverResponse)

	//If unmarshalling of the ServerResponse fails, this means that byte array got corrupted on network transmission
	if err != nil {
		return serverResponse, err
	}
	return serverResponse, nil
}

// DeserializePeerRequest converts a JSON byte array into a PeerRequest struct
func DeserializePeerRequest(b []byte) (data.PeerRequest, error) {
	var peerRequest data.PeerRequest
	err := json.Unmarshal(b, &peerRequest)

	if err != nil {
		return peerRequest, err
	}
	return peerRequest, nil
}


// DeserializePeerResponseData deserializes a byte array into PeerResponseHeader and response data
func DeserializePeerResponseData(b []byte) (data.PeerResponseHeader, string, error) {
	var peerResponse data.PeerResponseHeader

	// Find the end of the JSON object by counting braces
	jsonEnd := -1
	braceCount := 0
	for i, char := range b {
		if char == '{' {
			braceCount++
		} else if char == '}' {
			braceCount--
			if braceCount == 0 {
				jsonEnd = i + 1
				break
			}
		}
	}

	if jsonEnd == -1 {
		return peerResponse, "", json.Unmarshal(b, &peerResponse)
	}

	// Unmarshal the JSON header
	err := json.Unmarshal(b[:jsonEnd], &peerResponse)
	if err != nil {
		return peerResponse, "", err
	}

	// Extract the remaining data as string
	responseData := string(b[jsonEnd:])

	return peerResponse, responseData, nil
}

