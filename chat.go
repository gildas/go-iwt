package iwt

import (
	"fmt"
	"time"
)

// Chat describes a live chat
type Chat struct {
	ID                 string        `json:"chatID"`
	Participants       []Participant `json:"participants"`
	PollWaitSuggestion time.Duration `json:"pollWaitSuggestion"`
	Language           string        `json:"language"`
	DateFormat         string        `json:"dateFormat"`
	TimeFormat         string        `json:"timeFormat"`
}

func (chat *Chat) String() string {
	return chat.ID
}

// StartChatOptions defines the options when starting a chat
type StartChatOptions struct {
	SupportedContentTypes string            `json:"supportedContentTypes"`
	Participant           Participant       `json:"participant"`
	TranscriptRequired    bool              `json:"transcriptRequired"`
	EmailAddress          string            `json:"emailAddress,omitempty"`
	Language              string            `json:"language,omitempty"`
	QueueName             string            `json:"target"`
	QueueType             string            `json:"targettype"`
	Attributes            map[string]string `json:"attributes,omitempty"`
	RoutingContexts       []RoutingContext  `json:"routingContexts,omitempty"`
}

// RoutingContext defines the routing context when starting a chat (see IWT documentation)
type RoutingContext struct {
	Category string `json:"category"`
	Context  string `json:"context"`
}

type startChatResults struct {
	ID                 string `json:"chatID"`
	ParticipantID      string `json:"participantID"`
	PollWaitSuggestion int    `json:"pollWaitSuggestion"` // in ms => time.Duration
	DateFormat         string `json:"dateFormat"`
	TimeFormat         string `json:"timeFormat"`
	Status             Status `json:"status"`
	Version            int    `json:"cfgVer"`
}

// StartChat starts a chat
func (client *Client) StartChat(options StartChatOptions) (*Chat, error) {
	// Sanitizing options
	options.SupportedContentTypes = "text/plain" // only supported types so far...
	switch {
	case len(options.QueueType) == 0:
		options.QueueType = "Workgroup"
	case options.QueueType == "Station":
	case options.QueueType == "User":
	case options.QueueType == "Workgroup":
	default:
		return nil, fmt.Errorf("error.websvc.unknownEntity.invalidQueueType")
	}

	results := struct{Chat startChatResults `json:"chat"`}{}
	_, err := client.sendRequest(client.Context, &requestOptions{
		Path:    "/chat/start",
		Payload: options,
	}, &results)
	if err != nil {
		return nil, err
	}
	chat := Chat{
		ID:                 results.Chat.ID,
		Participants:       []Participant{Participant{ID: results.Chat.ParticipantID, Name: options.Participant.Name, State: "active"}},
		PollWaitSuggestion: time.Duration(results.Chat.PollWaitSuggestion) * time.Millisecond,
		Language:           options.Language,
		DateFormat:         results.Chat.DateFormat,
		TimeFormat:         results.Chat.TimeFormat,
	}
	return &chat, results.Chat.Status.AsError()
}