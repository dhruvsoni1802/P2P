package main

import (
	"P2P/common-helpers/data"
	"encoding/json"
)

// DeserializeAddStruct converts a JSON byte array into an AddStruct
func DeserializeAddStruct(b []byte) (data.AddStruct, error) {
	var addStruct data.AddStruct
	err := json.Unmarshal(b, &addStruct)

	//If unmarshalling of the AddStruct fails, this means that byte array got corrupted on network transmission
	if err != nil {
		return addStruct, err
	}
	return addStruct, nil
}

// DeserializeLookUpStruct converts a JSON byte array into a LookUpStruct
func DeserializeLookUpStruct(b []byte) (data.LookUpStruct, error) {
	var lookUpStruct data.LookUpStruct
	err := json.Unmarshal(b, &lookUpStruct)
	if err != nil {
		return lookUpStruct, err
	}
	return lookUpStruct, nil
}

// DeserializeListStruct converts a JSON byte array into a ListStruct
func DeserializeListStruct(b []byte) (data.ListStruct, error) {
	var listStruct data.ListStruct
	err := json.Unmarshal(b, &listStruct)
	if err != nil {
		return listStruct, err
	}
	return listStruct, nil
}
