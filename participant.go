package iwt

import (
	"encoding/json"
	"net/url"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// Participant defines a chat participant
type Participant struct {
	Type        string   `json:"participantType"`
	ID          string   `json:"participantID,omitempty"`
	Name        string   `json:"name"`
	Credentials string   `json:"credentials,omitempty"`
	Picture     *url.URL `json:"-"`
	State       string   `json:"state,omitempty"`
	Status      Status   `json:"status"`
}

var (
	// SystemParticipant is used for messages sent by PureConnect
	SystemParticipant = Participant{Type: "System", ID: "00000000-0000-0000-0000-000000000000", Name: "System", State: "active"}
)

// GetParticipant fetches a participant from a chat by its ID
func (chat *Chat) GetParticipant(id string) (*Participant, error) {
	log := chat.Logger.Scope("partyinfo")

	if len(chat.ID) == 0 || len(chat.Participants) == 0 || len(chat.Participants[0].ID) == 0 {
		log.Errorf("chat is not connected")
		return nil, StatusNotConnectedEntity
	}

	if id == SystemParticipant.ID {
		return &SystemParticipant, nil
	}

	log.Debugf("Requesting party information...")
	results := struct {
		Participant Participant `json:"partyInfo"`
	}{}
	_, err := chat.Client.post("/partyInfo/"+chat.Participants[0].ID,
		struct {
			ParticipantID string `json:"participantID"`
		}{id}, &results)
	if err != nil {
		return nil, err
	}
	results.Participant.ID = id
	return &results.Participant, results.Participant.Status.AsError()
}

// MarshalJSON encodes into JSON
func (participant Participant) MarshalJSON() ([]byte, error) {
	type surrogate Participant
	payload, err := json.Marshal(struct {
		surrogate
		P *core.URL `json:"photo,omitempty"`
	}{
		surrogate(participant),
		(*core.URL)(participant.Picture),
	})
	return payload, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON decodes JSON
func (participant *Participant) UnmarshalJSON(payload []byte) (err error) {
	type surrogate Participant
	var inner struct {
		surrogate
		Type string    `json:"type"`
		P    *core.URL `json:"photo,omitempty"`
	}
	if err = json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	*participant = Participant(inner.surrogate)
	participant.Picture = (*url.URL)(inner.P)
	return
}
