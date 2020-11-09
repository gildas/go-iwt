package iwt

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

// FileEvent describes the Text event
type FileEvent struct {
	SequenceNumber             int         `json:"sequenceNumber"`
	ConversationSequenceNumber int         `json:"conversationSequenceNumber"`
	Participant                Participant `json:"-"`
	ContentType                string      `json:"contentType"`
	Path                       string      `json:"value"`
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
func (event *FileEvent) UnmarshalJSON(payload []byte) (err error) {
	type surrogate FileEvent
	var inner struct {
		surrogate
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*event = FileEvent(inner.surrogate)

	// Capture the participant from the same payload
	if err = json.Unmarshal(payload, &event.Participant); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	return
}
