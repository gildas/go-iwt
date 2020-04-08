package iwt

import (
	"encoding/json"
	"net/url"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// URLEvent describes the Text event
type URLEvent struct {
	ParticipantID   string   `json:"participantID"`
	ParticipantName string   `json:"displayName"`
	SequenceNumber  int      `json:"sequenceNumber"`
	URL             *url.URL `json:"-"`
}

// GetType returns the type of this event
func (event URLEvent) GetType() string {
	return "url"
}

func (event URLEvent) String() string {
	return event.URL.String()
}

// MarshalJSON encodes into JSON
func (event URLEvent) MarshalJSON() ([]byte, error) {
	type surrogate URLEvent
	payload, err := json.Marshal(struct {
		surrogate
		Type string    `json:"type"`
		U    *core.URL `json:"value"`
	}{
		surrogate(event),
		event.GetType(),
		(*core.URL)(event.URL),
	})
	return payload, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON decodes JSON
func (event *URLEvent) UnmarshalJSON(payload []byte) (err error) {
	type surrogate URLEvent
	var inner struct {
		surrogate
		Type string    `json:"type"`
		U    *core.URL `json:"value"`
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*event = URLEvent(inner.surrogate)
	event.URL = (*url.URL)(inner.U)
	return
}
