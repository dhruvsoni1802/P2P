package main

import (
	"fmt"
	"log"
	"net"

	common_helpers "P2P/common-helpers"
	"P2P/common-helpers/data"
)

// sendErrorResponse sends an error response to the client
func sendErrorResponse(conn net.Conn, code int, phrase string) error {
	response := data.ServerResponse{
		Header: data.ServerResponseHeader{
			ResponseCode:             code,
			ResponsePhrase:           phrase,
			ServerApplicationVersion: ApplicationVersion,
		},
		Data: []data.ServerResponseData{},
	}

	serialized, err := SerializeServerResponse(response)
	if err != nil {
		return fmt.Errorf("error serializing response: %w", err)
	}

	log.Printf("Error response created: %s", string(serialized))
	serialized = append(serialized, '\n')
	_, err = conn.Write(serialized)
	log.Printf("Error response sent to client: %s", string(serialized))
	return err
}

// sendSuccessResponse sends a success response with data to the client
func sendSuccessResponse(conn net.Conn, responseData []data.ServerResponseData) error {
	response := data.ServerResponse{
		Header: data.ServerResponseHeader{
			ResponseCode:             StatusOK,
			ResponsePhrase:           "OK",
			ServerApplicationVersion: ApplicationVersion,
		},
		Data: responseData,
	}

	serialized, err := SerializeServerResponse(response)
	if err != nil {
		return fmt.Errorf("error serializing response: %w", err)
	}

	serialized = append(serialized, '\n')
	_, err = conn.Write(serialized)
	return err
}

// rfcExists checks if an RFC already exists in the index for a given client
func rfcExists(clientIP, rfcNumber, rfcTitle string) bool {
	rfcIndexMapMutex.RLock()
	defer rfcIndexMapMutex.RUnlock()

	for _, rfcInfo := range rfcIndexMap[clientIP] {
		if len(rfcInfo) == 2 && rfcInfo[0] == rfcNumber && rfcInfo[1] == rfcTitle {
			return true
		}
	}
	return false
}

// addRFCToIndex adds an RFC to the index for a given hostname
func addRFCToIndex(hostname, rfcNumber, rfcTitle string) {
	rfcIndexMapMutex.Lock()
	defer rfcIndexMapMutex.Unlock()

	// Initialize if needed
	if _, ok := rfcIndexMap[hostname]; !ok {
		rfcIndexMap[hostname] = make([][]string, 0)
	}

	rfcIndexMap[hostname] = append(rfcIndexMap[hostname], []string{rfcNumber, rfcTitle})
	log.Printf("Added RFC %s (%s) for host %s", rfcNumber, rfcTitle, hostname)
}

// peerExists checks if a peer already exists in the peer info map
func peerExists(clientIP string) bool {
	peerInfoMapMutex.RLock()
	defer peerInfoMapMutex.RUnlock()

	_, exists := peerInfoMap[clientIP]
	return exists
}

// addPeerInfo adds peer information to the peer info map
func addPeerInfo(hostname, uploadPort string) {
	peerInfoMapMutex.Lock()
	defer peerInfoMapMutex.Unlock()

	peerInfoMap[hostname] = uploadPort
	log.Printf("Added peer info for host %s on port %s", hostname, uploadPort)
}

// removePeerInfo removes peer information when a client disconnects
func removePeerInfo(clientAddr string) {
	peerInfoMapMutex.Lock()
	defer peerInfoMapMutex.Unlock()

	delete(peerInfoMap, clientAddr)
	log.Printf("Removed peer info for %s", clientAddr)
}

// removeRFCIndex removes RFC index when a client disconnects
func removeRFCIndex(clientAddr string) {
	rfcIndexMapMutex.Lock()
	defer rfcIndexMapMutex.Unlock()

	delete(rfcIndexMap, clientAddr)
	log.Printf("Removed RFC index for %s", clientAddr)
}

// handleAddRequest processes an ADD request from a client
func handleAddRequest(conn net.Conn, jsonData []byte) error {
	addStruct, err := DeserializeAddStruct(jsonData)
	if err != nil {
		log.Printf("Error deserializing AddStruct: %v", err)
		return sendErrorResponse(conn, StatusBadRequest, "Bad Request")
	}

	log.Printf("ADD request: RFC %s (%s) from %s on upload port %s with application version %s",
		addStruct.RFCNumber, addStruct.RFCTitle, addStruct.ClientIP, addStruct.ClientUploadPort, addStruct.ClientApplicationVersion)

	// Validate application version
	if addStruct.ClientApplicationVersion != ApplicationVersion {
		log.Printf("Version mismatch: client=%s, server=%s",
			addStruct.ClientApplicationVersion, ApplicationVersion)
		return sendErrorResponse(conn, StatusVersionNotSupported, "P2P-CI Version Not Supported")
	}
  
	// Check if RFC already exists 
	if rfcExists(addStruct.ClientIP, addStruct.RFCNumber, addStruct.RFCTitle) {
		log.Printf("RFC %s already exists for %s", addStruct.RFCNumber, addStruct.ClientIP)
		// Still send success response
		responseData := data.ServerResponseData{
			RFCNumber:        addStruct.RFCNumber,
			RFCTitle:         addStruct.RFCTitle,
			ClientIP:         addStruct.ClientIP,
			ClientUploadPort: addStruct.ClientUploadPort,
		}
		return sendSuccessResponse(conn, []data.ServerResponseData{responseData})
	}

	// Add RFC to index
	addRFCToIndex(addStruct.ClientIP, addStruct.RFCNumber, addStruct.RFCTitle)

	// Add peer info if not already present
	if !peerExists(addStruct.ClientIP) {
		addPeerInfo(addStruct.ClientIP, addStruct.ClientUploadPort)
	}

	// Send success response
	responseData := data.ServerResponseData{
		RFCNumber:        addStruct.RFCNumber,
		RFCTitle:         addStruct.RFCTitle,
		ClientIP:         addStruct.ClientIP,
		ClientUploadPort: addStruct.ClientUploadPort,
	}
	return sendSuccessResponse(conn, []data.ServerResponseData{responseData})
}

// handleLookupRequest processes a LOOKUP request from a client
func handleLookupRequest(conn net.Conn, jsonData []byte) error {
	lookUpStruct, err := DeserializeLookUpStruct(jsonData)
	if err != nil {
		log.Printf("Error deserializing LookUpStruct: %v", err)
		return sendErrorResponse(conn, StatusBadRequest, "Bad Request")
	}

	log.Printf("LOOKUP request: RFC %s (%s) from %s:%s",
		lookUpStruct.RFCNumber, lookUpStruct.RFCTitle,
		lookUpStruct.ClientIP, lookUpStruct.ClientUploadPort)

	// TODO: Implement lookup logic to search RFC index and return matching peers
	return nil
}

// handleListRequest processes a LIST request from a client
func handleListRequest(conn net.Conn, jsonData []byte) error {
	listStruct, err := DeserializeListStruct(jsonData)
	if err != nil {
		log.Printf("Error deserializing ListStruct: %v", err)
		return sendErrorResponse(conn, StatusBadRequest, "Bad Request")
	}

	log.Printf("LIST request from %s:%s",
		listStruct.ClientIP, listStruct.ClientUploadPort)

	//We create an empty array of ServerResponseData
	responseData := []data.ServerResponseData{}

	// Iterate through the complete rfcIndexMap
	for clientIP, rfcInfoArray := range rfcIndexMap {
			
		// Now iterate through each RFC pair for this clientIP
		for _, rfcInfo := range rfcInfoArray {
				
				// Now we do a lookup in peerInfoMap to get the upload port
				uploadPort, ok := peerInfoMap[clientIP]
				if !ok {
						log.Printf("Upload port not found for client %s", clientIP)
						continue
				}

				// Now we add the RFC information to the responseData
				responseData = append(responseData, data.ServerResponseData{
						RFCNumber: rfcInfo[0],
						RFCTitle: rfcInfo[1],
						ClientIP: clientIP,
						ClientUploadPort: uploadPort,
				})
		}
	}
	
	//Now we send the responseData to the client
	return sendSuccessResponse(conn, responseData)
}

// handleClientMessages listens for and processes messages from a client connection
func handleClientMessages(conn net.Conn, dedicatedPort string) {
	defer conn.Close()
	defer common_helpers.ReturnPort(dedicatedPort)
	defer removeRFCIndex(conn.RemoteAddr().String())
	defer removePeerInfo(conn.RemoteAddr().String())

	reader := common_helpers.NewMessageReader(conn)
	for {
		message, err := reader.ReadMessage()
		if err != nil {
			log.Printf("Error reading message from %s: %v", conn.RemoteAddr(), err)
			return
		}

		// Extract message type and payload
		if len(message) < 2 {
			log.Printf("Invalid message length from %s", conn.RemoteAddr())
			continue
		}

		structTypeInt := int(message[0])
		jsonData := message[1 : len(message)-1]

		// Route to appropriate handler
		var handleErr error
		switch structTypeInt {
		case common_helpers.AddStructIndex:
			handleErr = handleAddRequest(conn, jsonData)
		case common_helpers.LookupStructIndex:
			handleErr = handleLookupRequest(conn, jsonData)
		case common_helpers.ListStructIndex:
			handleErr = handleListRequest(conn, jsonData)
		default:
			log.Printf("Unknown message type %d from %s", structTypeInt, conn.RemoteAddr())
			continue
		}

		if handleErr != nil {
			log.Printf("Error handling request: %v", handleErr)
			return
		}
	}
}
