package main

import (
	"P2P/common-helpers/data"
	"encoding/json"
)

// SerializeAddStruct converts AddStruct into a JSON byte array
func SerializeAddStruct(addStruct data.AddStruct) ([]byte, error) {
	jsonData, err := json.Marshal(addStruct)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// SerializeLookUpStruct converts LookUpStruct into a JSON byte array
func SerializeLookUpStruct(lookUpStruct data.LookUpStruct) ([]byte, error) {
	jsonData, err := json.Marshal(lookUpStruct)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// SerializeListStruct converts ListStruct into a JSON byte array
func SerializeListStruct(listStruct data.ListStruct) ([]byte, error) {
	jsonData, err := json.Marshal(listStruct)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}
