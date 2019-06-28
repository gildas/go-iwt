package iwt

import (
	"fmt"
)

// ParticipantStateChangedEvent describes the ParticipantStateChanged event
type ParticipantStateChangedEvent struct {
	ParticipantID   string `json:"participantID"`
	ParticipantName string `json:"participantName"`
	SequenceNumber  int    `json:"sequenceNumber"`
	State           string `json:"state"`
}

// GetType returns the type of this event
func (event ParticipantStateChangedEvent) GetType() string {
	return "participantStateChanged"
}

func (event ParticipantStateChangedEvent) String() string {
	return fmt.Sprintf("Participant %s (%s) new state: %s", event.ParticipantName, event.ParticipantID, event.State)
}