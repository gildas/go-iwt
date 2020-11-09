package iwt

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

// TextEvent describes the Text event
type TextEvent struct {
	SequenceNumber             int         `json:"sequenceNumber"`
	ConversationSequenceNumber int         `json:"conversationSequenceNumber"`
	Participant                Participant `json:"-"`
	ContentType                string      `json:"contentType"`
	Text                       string      `json:"value"`
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
func (event *TextEvent) UnmarshalJSON(payload []byte) (err error) {
	type surrogate TextEvent
	var inner struct {
		surrogate
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*event = TextEvent(inner.surrogate)

	// Capture the participant from the same payload
	if err = json.Unmarshal(payload, &event.Participant); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	return
}
