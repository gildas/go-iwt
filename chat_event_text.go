package iwt

// TextEvent describes the Text event
type TextEvent struct {
	ParticipantID              string `json:"participantID"`
	ParticipantName            string `json:"participantName"`
	SequenceNumber             int    `json:"sequenceNumber"`
	ConversationSequenceNumber int    `json:"conversationSequenceNumber"`
	ContentType                string `json:"contentType"`
	Text                       string `json:"value"`
}

// GetType returns the type of this event
func (event TextEvent) GetType() string {
	return "text"
}

func (event TextEvent) String() string {
	return event.Text
}
