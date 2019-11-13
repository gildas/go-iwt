package iwt

import (
	"strings"
	"time"

	"github.com/gildas/go-core"
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

// IsWebUser tells if the given participantID is the customer (WebUser)
func (chat Chat) IsWebUser(participantID string) bool {
	return len(chat.Participants) > 0 && participantID == chat.Participants[0].ID
}

// StartChat starts a chat
// Chat Events will be sent to Chat.EventChan
func (client *Client) StartChat(options StartChatOptions) (*Chat, error) {
	log := client.Logger.Child("chat", "start")

	// Sanitizing options
	options.SupportedContentTypes = "text/plain" // only supported types so far...

	log.Debugf("Starting a Chat in %s", options.Queue.String())
	results := struct{Chat chatResponse `json:"chat"`}{}
	_, err := client.post("/chat/start",
		chatRequest{
			options.Queue.Name,
			options.Queue.Type,
			options,
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
		Logger:             client.Logger.Child("chat", "chat", "chat", results.Chat.ID),
	}
	chat.Logger.Infof("Chat created on queue %s with %s (%s)", chat.Queue, chat.Participants[0].Name, chat.Participants[0].ID)
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
	_, err := chat.Client.post("/chat/exit/" + chat.Participants[0].ID, nil, &results)
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
	_, err := chat.Client.post("/chat/reconnect", struct {ChatID string `json:"chatID"`}{chat.ID}, &results)
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
	_, err := chat.Client.post("/chat/sendMessage/" + chat.Participants[0].ID,
		struct {
			Message     string `json:"message"`
			ContentType string `json:"contentType"`
		}{text, contentType},
		&results)
	if err != nil {
		log.Errorf("Failed to send /chat/sendMessage request", err)
		return err
	}
	go chat.processEvents(results.Chat.Events)
	return results.Chat.Status.Param("id", chat.ID).AsError()
}

// GetFile download a file sent by an agent
func (chat *Chat) GetFile(path string) (reader *core.ContentReader, err error) {
	log := chat.Logger.Scope("getfile")
	if len(chat.ID) == 0 {
		log.Errorf("chat is not connected")
		return nil, StatusNotConnectedEntity
	}

	log.Debugf("Requesting file...")
	reader, err = chat.Client.get(strings.TrimPrefix(path, "/websvcs"), nil)
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
		log := chat.Logger.Scope("pollmessages")
		for {
			select {
			case <- chat.PollTicker.C:
				if len(chat.Participants) == 0 {
					log.Warnf("Chat has no participant...")
					chat.stopPollingMessages()
					chat.EventChan <- StopEvent{ChatID: chat.ID}
					chat.ID = ""
					return
				}
				if len(chat.Participants[0].ID) == 0 {
					log.Errorf("Chat first participant has no ID... (name=%s, state=%s)", chat.Participants[0].Name, chat.Participants[0].State)
					chat.stopPollingMessages()
					chat.EventChan <- StopEvent{ChatID: chat.ID}
					chat.ID = ""
					return
				}
				log.Debugf("Polling messages for Participant %s (%s) %s", chat.Participants[0].Name, chat.Participants[0].ID, chat.Participants[0].State)
				switch chat.Participants[0].State {
				case "disconnected":
					log.Infof("First participant disconnected, stopping chat")
					chat.EventChan <- StopEvent{ChatID: chat.ID}
					chat.stopPollingMessages()
					chat.ID = ""
					return
				case "active":
					results := struct{Chat chatResponse `json:"chat"`}{}
					_, err := chat.Client.get("/chat/poll/"+chat.Participants[0].ID, &results)
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
	log := chat.Logger.Scope("processevents")

	for _, event := range events {
		log.Record("event", event).Debugf("Emitting Event %s...", event.Event.GetType())
		switch event.Event.GetType() {
		case ParticipantStateChangedEvent{}.GetType():
			evt := event.Event.(ParticipantStateChangedEvent)
			if evt.State == "disconnected" {
				chat.EventChan <- StopEvent{ ChatID: chat.ID }
			} else {
				chat.EventChan <- event.Event
			}
		case TextEvent{}.GetType():
			evt := event.Event.(TextEvent)
			if evt.ParticipantType == "WebUser" && chat.IsWebUser(evt.ParticipantID) {
				log.Debugf("This is an echo of a message sent by the WebUser, ignoring it")
				continue
			} else {
				chat.EventChan <- event.Event
			}
		default:
			chat.EventChan <- event.Event
		}
	}
}