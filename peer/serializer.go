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


func SerializeLookUpStruct(lookUpStruct data.LookUpStruct) ([]byte, error) {
	json, err := json.Marshal(lookUpStruct)
	if err != nil {
		return nil, err
	}
	return json, nil
}

func SerializeListStruct(listStruct data.ListStruct) ([]byte, error) {
	json, err := json.Marshal(listStruct)
	if err != nil {
		return nil, err
	}
	return json, nil
}