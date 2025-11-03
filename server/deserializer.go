package main

import (
	"P2P/common-helpers/data"
	"encoding/json"
)

//Function to convert a byte array into an AddStruct
func DeserializeAddStruct(b []byte) (data.AddStruct, error) {
	var addStruct data.AddStruct
	err := json.Unmarshal(b, &addStruct)
	if err != nil {
		return addStruct, err
	}
	return addStruct, nil
}