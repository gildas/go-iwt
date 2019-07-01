package iwt

import (
	"encoding/json"
)

// TextEvent describes the Text event
type TextEvent struct {
	ParticipantID              string `json:"participantID"`
	ParticipantName            string `json:"displayName"`
	ParticipantType            string `json:"participantType"`
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

// MarshalJSON encodes into JSON
func (event TextEvent) MarshalJSON() ([]byte, error) {
	type surrogate TextEvent
	return json.Marshal(struct {
		surrogate
		Type string    `json:"type"`
	}{
		surrogate(event),
		event.GetType(),
	})
}