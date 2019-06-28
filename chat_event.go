package iwt

import (
	"encoding/json"
	"fmt"
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
	header := struct {Type string `json:"type"`}{}
	if err = json.Unmarshal(payload, &header); err != nil { return }
	switch header.Type {
	case ParticipantStateChangedEvent{}.GetType():
		var inner ParticipantStateChangedEvent
		if err = json.Unmarshal(payload, &inner); err != nil { return }
		wrapper.Event = ParticipantStateChangedEvent(inner)
	case TextEvent{}.GetType():
		var inner TextEvent
		if err = json.Unmarshal(payload, &inner); err != nil { return }
		wrapper.Event = TextEvent(inner)
	default:
		return fmt.Errorf("Unsupported ChatEvent %s", header.Type)
	}
	return nil
}