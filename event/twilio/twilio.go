package twilio

type Event struct {
	Event string `json:"event"`
}

type Start struct {
	SequenceNumber string     `json:"sequenceNumber"`
	Start          *StartInfo `json:"start"`
	StreamSid      string     `json:"streamSid"`
}

type StartInfo struct {
	AccountSid       string         `json:"accountSid"`
	CallSid          string         `json:"callSid"`
	Tracks           []string       `json:"tracks"`
	CustomParameters map[string]any `json:"customParameters"`
}

type Media struct {
	SequenceNumber string     `json:"sequenceNumber"`
	Media          *MediaInfo `json:"media"`
	StreamSid      string     `json:"streamSid"`
}

type MediaInfo struct {
	Track     string `json:"track"`
	Chunk     string `json:"chunk"`
	Timestamp string `json:"timestamp"`
	Payload   string `json:"payload"`
}
