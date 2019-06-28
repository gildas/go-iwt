package iwt

import (
	"errors"
	"strings"
)

// QueueType defines the type of Queue
type QueueType int
const (
	// StationQueue queue type
	StationQueue QueueType = iota
	// UserQueue queue type
	UserQueue
	// WorkgroupQueue queue type
	WorkgroupQueue
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
		*queueType = WorkgroupQueue
	case "user":
		*queueType = UserQueue
	case "station":
		*queueType = StationQueue
	default:
		return errors.New("Invalid QueueType: " + unquoted)
	}
	return nil
}

func (queueType QueueType) String() string {
	return []string{"Station", "User", "Workgroup"}[queueType]
}