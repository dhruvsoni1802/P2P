package data

// PeerResponse represents the response from a peer
type PeerResponseHeader struct {
	PeerApplicationVersion    string
	Status                    int
	Phrase                    string
	CurrentDateandTime        string
	OS                        string
	LastModifiedDateandTime   string
	ContentLength             string
	ContentType               string
}

