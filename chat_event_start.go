package iwt

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

// StartEvent describes the Start event
type StartEvent struct {
	ChatID       string        `json:"chatID"`
	Participants []Participant `json:"participants"`
	Guest        Participant   `json:"guest"` // used to store the id of the guest on their platform (LINE, KKT, etc)
	Language     string        `json:"language"`
	DateFormat   string        `json:"dateFormat"`
	TimeFormat   string        `json:"timeFormat"`
}

// GetType returns the type of this event
func (event StartEvent) GetType() string {
	return "start"
}

func (event StartEvent) String() string {
	return event.ChatID
}

// MarshalJSON encodes into JSON
func (event StartEvent) MarshalJSON() ([]byte, error) {
	type surrogate StartEvent
	payload, err := json.Marshal(struct {
		surrogate
		Type string `json:"type"`
	}{
		surrogate(event),
		event.GetType(),
	})
	return payload, errors.JSONMarshalError.Wrap(err)
}
