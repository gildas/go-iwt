package iwt

// Participant defines a chat participant
type Participant struct {
	ID          string `json:"participantID,omitempty"`
	Name        string `json:"name"`
	Credentials string `json:"credentials,omitempty"`
	Picture     string `json:"photo,omitempty"`
	State       string `json:"state,omitempty"`
	Status      Status `json:"status"`
}

// GetParticipant fetches a participant from a chat by its ID
func (chat *Chat) GetParticipant(id string) (*Participant, error) {
	log := chat.Logger.Scope("partyinfo")

	if len(chat.ID) == 0 || len(chat.Participants) == 0 || len(chat.Participants[0].ID) == 0 {
		log.Errorf("chat is not connected")
		return nil, StatusNotConnectedEntity
	}

	log.Debugf("Requesting party information...")
	results := struct{Participant Participant `json:"partyInfo"`}{}
	_, _, err := chat.Client.sendRequest(chat.Client.Context, &requestOptions{
		Path: "/partyInfo/" + chat.Participants[0].ID,
		Payload: struct {ParticipantID string `json:"participantID"`}{id},
	}, &results)
	if err != nil {
		return nil, err
	}
	results.Participant.ID = id
	return &results.Participant, results.Participant.Status.AsError()
}