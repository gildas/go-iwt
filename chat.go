package iwt

import (
	"io"
	"net/http"
	"time"

	"github.com/gildas/go-logger"
)

// Chat describes a live chat
type Chat struct {
	ID                 string         `json:"chatID"`
	Queue              *Queue         `json:"queue"`
	Participants       []Participant  `json:"participants"`
	PollWaitSuggestion time.Duration  `json:"pollWaitSuggestion"`
	Language           string         `json:"language"`
	DateFormat         string         `json:"dateFormat"`
	TimeFormat         string         `json:"timeFormat"`
	Client             *Client        `json:"-"`
	Logger             *logger.Logger `json:"-"`
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
	QueueType             QueueType         `json:"targettype"`
	Attributes            map[string]string `json:"attributes,omitempty"`
	RoutingContexts       []RoutingContext  `json:"routingContexts,omitempty"`
}

// RoutingContext defines the routing context when starting a chat (see IWT documentation)
type RoutingContext struct {
	Category string `json:"category"`
	Context  string `json:"context"`
}

type chatResponse struct {
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
	log := client.Logger.Topic("chat").Scope("start").Child()

	// Sanitizing options
	options.SupportedContentTypes = "text/plain" // only supported types so far...
	queue := &Queue{Name: options.QueueName, Type: options.QueueType}

	log.Debugf("Starting a Chat in %s", queue)
	results := struct{Chat chatResponse `json:"chat"`}{}
	_, _, err := client.sendRequest(client.Context, &requestOptions{
		Path:    "/chat/start",
		Payload: options,
	}, &results)
	if err != nil {
		return nil, err
	}
	chat := Chat{
		ID:                 results.Chat.ID,
		Queue:              queue,
		Participants:       []Participant{Participant{ID: results.Chat.ParticipantID, Name: options.Participant.Name, State: "active"}},
		PollWaitSuggestion: time.Duration(results.Chat.PollWaitSuggestion) * time.Millisecond,
		Language:           options.Language,
		DateFormat:         results.Chat.DateFormat,
		TimeFormat:         results.Chat.TimeFormat,
		Client:             client,
		Logger:             log.Record("chat", results.Chat.ID).Child(),
	}
	// Start the polling go subroutine
	// return a chan that will receive Event objects (TBD)
	return &chat, results.Chat.Status.AsError()
}

// Stop stops the current chat
func (chat *Chat) Stop() error {
	log := chat.Logger.Scope("stop")

	if len(chat.ID) == 0 {
		log.Debugf("Chat is already stopped")
		return nil
	}

	log.Debugf("Stopping chat...")
	results := struct{Chat chatResponse `json:"chat"`}{}
	_, _, err := chat.Client.sendRequest(chat.Client.Context, &requestOptions{
		Method: http.MethodPost,
		Path:   "/chat/exit/" + chat.ID,
	}, &results)
	if err != nil {
		log.Errorf("Failed to send /chat/exit request", err)
		return err
	}
	// TODO: we emit chatstoppedevent on the chan whatever happened
	if results.Chat.Status.IsOK() || results.Chat.Status.IsA(StatusUnknownEntitySession) {
		chat.ID = ""
		return nil
	}
	return results.Chat.Status.AsError()
}

// SendMessage sends a message to the chat
func (chat *Chat) SendMessage(text, contentType string) error {
	log := chat.Logger.Scope("sendmessage")
	if len(chat.ID) == 0 {
		log.Errorf("chat is not connected")
		return StatusNotConnectedEntity
	}
	if len(contentType) == 0 {
		contentType = "text/plain"
	}

	log.Debugf("Sending %s message...", contentType)
	results := struct{Chat chatResponse `json:"chat"`}{}
	_, _, err := chat.Client.sendRequest(chat.Client.Context, &requestOptions{
		Method: http.MethodPost,
		Path:   "/chat/sendMessage/" + chat.ID,
		Payload: struct {
			Message     string `json:"message"`
			ContentType string `json:"contentType"`
		}{text, contentType},
	}, &results)
	if err != nil {
		log.Errorf("Failed to send /chat/sendMessage request", err)
		return err
	}
	return results.Chat.Status.Param("id", chat.ID).AsError()
}

// GetFile download a file sent by an agent
func (chat *Chat) GetFile(filepath string) (contentType string, reader io.ReadCloser, err error) {
	log := chat.Logger.Scope("stop")
	if len(chat.ID) == 0 {
		log.Errorf("chat is not connected")
		return "", nil, StatusNotConnectedEntity
	}

	log.Debugf("Rquesting file...")
	reader, contentType, err = chat.Client.sendRequest(chat.Client.Context, &requestOptions{
		Path:   "/chat/sendMessage/" + chat.ID,
		Accept: "application/octet-stream",
	}, nil)
	if err != nil {
		log.Errorf("Failed to send /chat/sendMessage request", err)
		return
	}
	return
}