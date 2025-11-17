package data

// ListStruct represents the data structure for listing all RFCs
type ListStruct struct {
	ClientIP                 string `json:"Client_IP"`
	ClientUploadPort         string `json:"Client_Upload_Port"`
	ClientApplicationVersion string `json:"Client_Application_Version"`
}
