package iwt

import (
	"fmt"
)

// Status defines the status of a queue, chat, IWT request
type Status struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

var (
	StatusUnknownEntitySession = Status{"failure", "error.websvc.unknownEntity.session"}
)

// IsOK tells if the status is a success
func (status Status) IsOK() bool {
	return status.Type == "success"
}

// IsA tells if the status is the same as the given status
func (status Status) IsA(ref Status) bool {
	return status.Type == ref.Type && status.Reason == ref.Reason
}

// AsError converts a status to a GO error
func (status Status) AsError() error {
	if status.IsOK() {
		return nil
	}
	return fmt.Errorf(status.Reason)
}

func (status Status) Error() string {
	if status.IsOK() {
		return ""
	}
	return status.Reason
}