package iwt

import (
	"encoding/json"

	"github.com/gildas/go-errors"
)

// ChatEvent defines a Chat Event
type ChatEvent interface {
	GetType() string
	String() string
}

// ChatEventWrapper is used to Un/Marshal ChatEvent Objects
type ChatEventWrapper struct {
	Event ChatEvent
}

// UnmarshalJSON decodes JSON
func (wrapper *ChatEventWrapper) UnmarshalJSON(payload []byte) (err error) {
	header := struct {
		Type string `json:"type"`
	}{}
	if err = json.Unmarshal(payload, &header); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	switch header.Type {
	case FileEvent{}.GetType():
		var inner FileEvent
		if err = json.Unmarshal(payload, &inner); err != nil {
			return errors.JSONUnmarshalError.Wrap(err)
		}
		wrapper.Event = FileEvent(inner)
	case ParticipantStateChangedEvent{}.GetType():
		var inner ParticipantStateChangedEvent
		if err = json.Unmarshal(payload, &inner); err != nil {
			return errors.JSONUnmarshalError.Wrap(err)
		}
		wrapper.Event = ParticipantStateChangedEvent(inner)
	case StartEvent{}.GetType():
		var inner StartEvent
		if err = json.Unmarshal(payload, &inner); err != nil {
			return errors.JSONUnmarshalError.Wrap(err)
		}
		wrapper.Event = StartEvent(inner)
	case StopEvent{}.GetType():
		var inner StopEvent
		if err = json.Unmarshal(payload, &inner); err != nil {
			return errors.JSONUnmarshalError.Wrap(err)
		}
		wrapper.Event = StopEvent(inner)
	case TextEvent{}.GetType():
		var inner TextEvent
		if err = json.Unmarshal(payload, &inner); err != nil {
			return errors.JSONUnmarshalError.Wrap(err)
		}
		wrapper.Event = TextEvent(inner)
	case TypingIndicatorEvent{}.GetType():
		var inner TypingIndicatorEvent
		if err = json.Unmarshal(payload, &inner); err != nil {
			return errors.JSONUnmarshalError.Wrap(err)
		}
		wrapper.Event = TypingIndicatorEvent(inner)
	case URLEvent{}.GetType():
		var inner URLEvent
		if err = json.Unmarshal(payload, &inner); err != nil {
			return errors.JSONUnmarshalError.Wrap(err)
		}
		wrapper.Event = URLEvent(inner)
	default:
		return errors.Unsupported.With("Chat Event", header.Type).WithStack()
	}
	return nil
}
