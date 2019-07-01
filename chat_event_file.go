package iwt

import (
	"encoding/json"
)

// FileEvent describes the Text event
type FileEvent struct {
	ParticipantID              string `json:"participantID"`
	ParticipantName            string `json:"displayName"`
	ParticipantType            string `json:"participantType"`
	SequenceNumber             int    `json:"sequenceNumber"`
	ConversationSequenceNumber int    `json:"conversationSequenceNumber"`
	ContentType                string `json:"contentType"`
	Path                       string `json:"value"`
}

// GetType returns the type of this event
func (event FileEvent) GetType() string {
	return "file"
}

func (event FileEvent) String() string {
	return event.Path
}

// MarshalJSON encodes into JSON
func (event FileEvent) MarshalJSON() ([]byte, error) {
	type surrogate FileEvent
	return json.Marshal(struct {
		surrogate
		Type string    `json:"type"`
	}{
		surrogate(event),
		event.GetType(),
	})
}