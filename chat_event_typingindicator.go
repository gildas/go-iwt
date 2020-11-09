package iwt

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

// TypingIndicatorEvent describes the TypingIndicator event
type TypingIndicatorEvent struct {
	SequenceNumber int         `json:"sequenceNumber"`
	Participant    Participant `json:"-"`
	Typing         bool        `json:"value"`
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
	payload, err := json.Marshal(struct {
		surrogate
		Type string `json:"type"`
	}{
		surrogate(event),
		event.GetType(),
	})
	return payload, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON decodes JSON
func (event *TypingIndicatorEvent) UnmarshalJSON(payload []byte) (err error) {
	type surrogate TypingIndicatorEvent
	var inner struct {
		surrogate
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*event = TypingIndicatorEvent(inner.surrogate)

	// Capture the participant from the same payload
	if err = json.Unmarshal(payload, &event.Participant); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	return
}
