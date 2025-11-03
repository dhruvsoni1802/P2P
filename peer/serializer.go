package main

import (
	"P2P/common-helpers/data"
	"encoding/json"
)

//Function to convert AddStruct into a byte array
func SerializeAddStruct(addStruct data.AddStruct) ([]byte, error) {
	json, err := json.Marshal(addStruct)
	if err != nil {
		return nil, err
	}
	return json, nil
}

