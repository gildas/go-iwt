package iwt

import (
	"encoding/json"
)

// TypingIndicatorEvent describes the TypingIndicator event
type TypingIndicatorEvent struct {
	ParticipantID  string `json:"participantID"`
	SequenceNumber int    `json:"sequenceNumber"`
	ContentType    string `json:"contentType"`
	Typing         bool   `json:"value"`
}

// GetType returns the type of this event
func (event TypingIndicatorEvent) GetType() string {
	return "typingIndicator"
}

func (event TypingIndicatorEvent) String() string {
	if event.Typing {
		return "typing"
	}
	return "not typing"
}

// MarshalJSON encodes into JSON
func (event TypingIndicatorEvent) MarshalJSON() ([]byte, error) {
	type surrogate TypingIndicatorEvent
	return json.Marshal(struct {
		surrogate
		Type string    `json:"type"`
	}{
		surrogate(event),
		event.GetType(),
	})
}