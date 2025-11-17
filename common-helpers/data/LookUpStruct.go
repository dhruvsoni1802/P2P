package data

// LookUpStruct represents the data structure for looking up RFC information
type LookUpStruct struct {
	RFCNumber                string `json:"RFC_Number"`
	RFCTitle                 string `json:"RFC_Title"`
	ClientIP                 string `json:"Client_IP"`
	ClientUploadPort         string `json:"Client_Upload_Port"`
	ClientApplicationVersion string `json:"Client_Application_Version"`
}
