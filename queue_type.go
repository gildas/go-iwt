package iwt

import (
	"errors"
	"strings"
)

// QueueType defines the type of Queue
type QueueType int
const (
	// Station queue type
	Station QueueType = iota
	// User queue type
	User
	// Workgroup queue type
	Workgroup
)

// MarshalJSON encodes JSON
func (queueType QueueType) MarshalJSON() ([]byte, error) {
	return []byte(queueType.String()), nil
}

// UnmarshalJSON decodes JSON
func (queueType *QueueType) UnmarshalJSON(payload []byte) (err error) {
	unquoted := strings.TrimSpace(strings.Replace(string(payload), `"`, ``, -1))
	switch strings.ToLower(unquoted) {
	case "", "workgroup":
		*queueType = Workgroup
	case "user":
		*queueType = User
	case "station":
		*queueType = Station
	default:
		return errors.New("Invalid QueueType: " + unquoted)
	}
	return nil
}

func (queueType QueueType) String() string {
	return []string{"Station", "User", "Workgroup"}[queueType]
}