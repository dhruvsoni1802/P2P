package data

type ServerResponseHeader struct {
	Response_Code int
	Response_Phrase string
	Server_Application_Version string
}
type ServerResponseData struct {
	RFC_Number string
	RFC_Title string
	Client_IP string
	Client_Upload_Port string
}

type ServerResponse struct {
	Header ServerResponseHeader
	Data []ServerResponseData
}