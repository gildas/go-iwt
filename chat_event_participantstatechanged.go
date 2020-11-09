package iwt

import (
	"encoding/json"
	"fmt"

	"github.com/gildas/go-errors"
)

// ParticipantStateChangedEvent describes the ParticipantStateChanged event
type ParticipantStateChangedEvent struct {
	SequenceNumber int         `json:"sequenceNumber"`
	Participant    Participant `json:"-"`
}

// GetType returns the type of this event
func (event ParticipantStateChangedEvent) GetType() string {
	return "participantStateChanged"
}

func (event ParticipantStateChangedEvent) String() string {
	return fmt.Sprintf("Participant %s (%s) new state: %s", event.Participant.Name, event.Participant.ID, event.Participant.State)
}

// MarshalJSON encodes into JSON
func (event ParticipantStateChangedEvent) MarshalJSON() ([]byte, error) {
	type surrogate ParticipantStateChangedEvent
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
func (event *ParticipantStateChangedEvent) UnmarshalJSON(payload []byte) (err error) {
	type surrogate ParticipantStateChangedEvent
	var inner struct {
		surrogate
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*event = ParticipantStateChangedEvent(inner.surrogate)

	// Capture the participant from the same payload
	if err = json.Unmarshal(payload, &event.Participant); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	return
}
