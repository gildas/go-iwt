package iwt

import (
	"encoding/json"
)

// StopEvent describes the Text event
type StopEvent struct {
	ChatID string `json:"chatID"`
}

// GetType returns the type of this event
func (event StopEvent) GetType() string {
	return "stop"
}

func (event StopEvent) String() string {
	return "stop"
}

// MarshalJSON encodes into JSON
func (event StopEvent) MarshalJSON() ([]byte, error) {
	type surrogate StopEvent
	return json.Marshal(struct {
		surrogate
		Type string    `json:"type"`
	}{
		surrogate(event),
		event.GetType(),
	})
}
