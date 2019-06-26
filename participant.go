package iwt

// Participant defines a chat participant
type Participant struct {
	ID          string `json:"participantID,omitempty"`
	Name        string `json:"name"`
	Credentials string `json:"credentials,omitempty"`
	State       string `json:"state,omitempty"`
}