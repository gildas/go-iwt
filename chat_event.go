package iwt

import (
	"encoding/json"
	"reflect"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
)

// ChatEvent defines a Chat Event
type ChatEvent interface {
	GetType() string
	String() string
}

// chatEventWrapper is used to Un/Marshal ChatEvent Objects
type chatEventWrapper struct {
	Event ChatEvent
}

// UnmarshalChatEvent unmarshals a JSON payload into a ChatEvent
func UnmarshalChatEvent(payload []byte) (ChatEvent, error) {
	wrapper := chatEventWrapper{}
	if err := json.Unmarshal(payload, &wrapper); err != nil {
		return nil, errors.JSONUnmarshalError.Wrap(err)
	}
	return wrapper.Event, nil
}

// UnmarshalJSON decodes JSON
func (wrapper *chatEventWrapper) UnmarshalJSON(payload []byte) (err error) {
	header := struct {
		Type string `json:"type"`
	}{}
	if err = json.Unmarshal(payload, &header); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}

	var value ChatEvent

	registry := core.TypeRegistry{}.Add(
		FileEvent{},
		ParticipantStateChangedEvent{},
		StartEvent{},
		StopEvent{},
		TextEvent{},
		TypingIndicatorEvent{},
		URLEvent{},
	)
	if valueType, found := registry[header.Type]; found {
		value = reflect.New(valueType).Interface().(ChatEvent)
	} else {
		return errors.JSONUnmarshalError.Wrap(errors.Unsupported.With("type", header.Type))
	}
	if err = json.Unmarshal(payload, &value); err != nil {
		return
	}
	wrapper.Event = value
	return nil
}
