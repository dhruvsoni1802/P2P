package data

// ServerResponseHeader represents the header of the server response
type ServerResponseHeader struct {
	ResponseCode             int    `json:"Response_Code"`
	ResponsePhrase           string `json:"Response_Phrase"`
	ServerApplicationVersion string `json:"Server_Application_Version"`
}

// ServerResponseData represents individual RFC data in the server response
type ServerResponseData struct {
	RFCNumber        string `json:"RFC_Number"`
	RFCTitle         string `json:"RFC_Title"`
	ClientIP         string `json:"Client_IP"`
	ClientUploadPort string `json:"Client_Upload_Port"`
}

// ServerResponse represents the complete server response structure
type ServerResponse struct {
	Header ServerResponseHeader
	Data   []ServerResponseData
}
