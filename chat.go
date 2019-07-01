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
	Guest              Participant    `json:"guest"`            // used to store the id of the guest on their platform (LINE, KKT, etc)
	PollWaitSuggestion time.Duration  `json:"pollWaitSuggestion"`
	Language           string         `json:"language"`
	DateFormat         string         `json:"dateFormat"`
	TimeFormat         string         `json:"timeFormat"`
	EventChan          chan ChatEvent `json:"-"`
	PollTicker         *time.Ticker   `json:"-"`
	Client             *Client        `json:"-"`
	Logger             *logger.Logger `json:"-"`
}

func (chat *Chat) String() string {
	return chat.ID
}

// StartChatOptions defines the options when starting a chat
type StartChatOptions struct {
	Queue                 *Queue            `json:"-"`
	Guest                 Participant       `json:"participant"`
	Language              string            `json:"language,omitempty"`
	EmailAddress          string            `json:"emailAddress,omitempty"`
	SupportedContentTypes string            `json:"supportedContentTypes"`
	TranscriptRequired    bool              `json:"transcriptRequired"`
	Attributes            map[string]string `json:"attributes,omitempty"`
	RoutingContexts       []RoutingContext  `json:"routingContexts,omitempty"`
}

// RoutingContext defines the routing context when starting a chat (see IWT documentation)
type RoutingContext struct {
	Category string `json:"category"`
	Context  string `json:"context"`
}

type chatRequest struct {
	QueueName        string    `json:"target"`
	QueueType        QueueType `json:"targettype"`
	StartChatOptions
}

type chatResponse struct {
	ID                 string             `json:"chatID"`
	ParticipantID      string             `json:"participantID"`
	PollWaitSuggestion int                `json:"pollWaitSuggestion,omitempty"` // in ms => time.Duration
	DateFormat         string             `json:"dateFormat,omitempty"`
	TimeFormat         string             `json:"timeFormat,omitempty"`
	Events             []ChatEventWrapper `json:"events"`
	Status             Status             `json:"status"`
	Version            int                `json:"cfgVer"`
}

// StartChat starts a chat
// Chat Events will be sent to Chat.EventChan
func (client *Client) StartChat(options StartChatOptions) (*Chat, error) {
	log := client.Logger.Topic("chat").Scope("start").Child()

	// Sanitizing options
	options.SupportedContentTypes = "text/plain" // only supported types so far...

	log.Debugf("Starting a Chat in %s", options.Queue.String())
	results := struct{Chat chatResponse `json:"chat"`}{}
	_, _, err := client.sendRequest(client.Context, &requestOptions{
		Path:    "/chat/start",
		Payload: chatRequest{
			options.Queue.Name,
			options.Queue.Type,
			options,
		},
	}, &results)
	if err != nil {
		return nil, err
	}
	if results.Chat.PollWaitSuggestion < 1000 {
		results.Chat.PollWaitSuggestion = 1000
	}
	chat := Chat{
		ID:                 results.Chat.ID,
		Queue:              options.Queue,
		Participants:       []Participant{Participant{ID: results.Chat.ParticipantID, Name: options.Guest.Name, State: "active"}},
		Guest:              options.Guest,
		PollWaitSuggestion: time.Duration(results.Chat.PollWaitSuggestion) * time.Millisecond,
		Language:           options.Language,
		DateFormat:         results.Chat.DateFormat,
		TimeFormat:         results.Chat.TimeFormat,
		EventChan:          make(chan ChatEvent),
		Client:             client,
		Logger:             log.Record("chat", results.Chat.ID).Child(),
	}
	// Start the polling go subroutine
	chat.startPollingMessages()
	return &chat, results.Chat.Status.AsError()
}

// Stop stops the current chat
func (chat *Chat) Stop() error {
	log := chat.Logger.Scope("stop")

	if len(chat.ID) == 0 || len(chat.Participants) == 0 || len(chat.Participants[0].ID) == 0 {
		log.Debugf("Chat is already stopped")
		return nil
	}

	log.Debugf("Stopping chat...")
	results := struct{Chat chatResponse `json:"chat"`}{}
	_, _, err := chat.Client.sendRequest(chat.Client.Context, &requestOptions{
		Method: http.MethodPost,
		Path:   "/chat/exit/" + chat.Participants[0].ID,
	}, &results)
	if err != nil {
		log.Errorf("Failed to send /chat/exit request", err)
		return err
	}
	chat.stopPollingMessages()
	chat.EventChan <- StopEvent{ChatID: chat.ID}
	if results.Chat.Status.IsOK() || results.Chat.Status.IsA(StatusUnknownEntitySession) {
		chat.ID = ""
		return nil
	}
	return results.Chat.Status.AsError()
}

// Reconnect reconnects the current chat to another server (Switchover event, e.g.)
func (chat *Chat) Reconnect() error {
	log := chat.Logger.Scope("reconnect")

	chat.stopPollingMessages()
	chat.Client.NextAPIEndpoint()
	log.Debugf("Reconnecting chat to %s...", chat.Client.CurrentAPIEndpoint())
	results := struct{Chat chatResponse `json:"chat"`}{}
	_, _, err := chat.Client.sendRequest(chat.Client.Context, &requestOptions{
		Method:  http.MethodPost,
		Path:    "/chat/reconnect",
		Payload: struct {ChatID string `json:"chatID"`}{chat.ID},
	}, &results)
	if err != nil {
		log.Errorf("Failed to send /chat/exit request", err)
		return err
	}
	chat.processEvents(results.Chat.Events)
	chat.startPollingMessages()
	return results.Chat.Status.Param("id", chat.ID).AsError()
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
		Path:   "/chat/sendMessage/" + chat.Participants[0].ID,
		Payload: struct {
			Message     string `json:"message"`
			ContentType string `json:"contentType"`
		}{text, contentType},
	}, &results)
	if err != nil {
		log.Errorf("Failed to send /chat/sendMessage request", err)
		return err
	}
	go chat.processEvents(results.Chat.Events)
	return results.Chat.Status.Param("id", chat.ID).AsError()
}

// GetFile download a file sent by an agent
func (chat *Chat) GetFile(filepath string) (contentType string, reader io.ReadCloser, err error) {
	log := chat.Logger.Scope("stop")
	if len(chat.ID) == 0 {
		log.Errorf("chat is not connected")
		return "", nil, StatusNotConnectedEntity
	}

	log.Debugf("Requesting file...")
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

func (chat *Chat) startPollingMessages() {
	if chat.PollTicker != nil {
		chat.stopPollingMessages()
	}
	chat.Logger.Scope("pollmessages").Infof("Polling messages every %s", chat.PollWaitSuggestion)
	chat.PollTicker = time.NewTicker(chat.PollWaitSuggestion)

	go func() {
		log := chat.Logger.Scope("pollmessages").Child()
		for {
			select {
			case now := <- chat.PollTicker.C:
				log.Debugf("Polling messages at %s", now)
				if len(chat.Participants) == 0 {
					log.Infof("Chat has no participant...")
					chat.stopPollingMessages()
					chat.ID = ""
					return
				}
				switch chat.Participants[0].State {
				case "disconnected":
					log.Infof("First participant disconnected, stopping chat")
					// Emit ChatStoppedEvent (parm: chat.ID)
					chat.stopPollingMessages()
					chat.ID = ""
					return
				case "active":
					results := struct{Chat chatResponse `json:"chat"`}{}
					_, _, err := chat.Client.sendRequest(chat.Client.Context, &requestOptions{
						Path: "/chat/poll/"+chat.Participants[0].ID,
					}, &results)
					if err == StatusUnavailableService.AsError() && len(chat.Client.APIEndpoints) > 1 {
						log.Warnf("A Switchover happened!")
						chat.Reconnect()
						continue
					}
					if err != nil {
						log.Errorf("Failed to send /chat/poll request", err)
						continue
					}
					if results.Chat.Status.IsA(StatusUnknownEntitySession) {
						log.Warnf("Zombie Chat, stopping it")
						chat.stopPollingMessages()
						chat.ID = ""
						return
					}
					if !results.Chat.Status.IsOK() {
						log.Errorf("Results contains an error", results.Chat.Status.AsError())
						continue
					}
					chat.processEvents(results.Chat.Events)
				default:
					log.Warnf("Unsupported state %s for participant %s (%s)", chat.Participants[0].State, chat.Participants[0].Name, chat.Participants[0].ID)
				}
			}
		}
	}()
}

func (chat *Chat) stopPollingMessages() {
	if chat.PollTicker != nil {
		chat.Logger.Scope("pollmessages").Debugf("stopping polling messages")
		chat.PollTicker.Stop()
		chat.Logger.Scope("pollmessages").Infof("stopped polling messages")
	}
	chat.PollTicker = nil
}

func (chat *Chat) processEvents(events []ChatEventWrapper) {
	log := chat.Logger.Scope("processevents").Child()

	for _, event := range events {
		log.Record("event", event).Debugf("Emitting Event %s...", event.Event.GetType())
		chat.EventChan <- event.Event
	}
}