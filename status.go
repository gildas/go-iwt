package iwt

import (
	"strings"
	"fmt"
)

// Status defines the status of a queue, chat, IWT request
type Status struct {
	Type   string                 `json:"type"`
	Reason string                 `json:"reason"`
	Params map[string]interface{} `json:"params"`
}

var (
	StatusUnknownEntitySession = Status{"failure", "error.websvc.unknownEntity.session", nil}
	StatusNotConnectedEntity   = Status{"failure", "error.websvc.entity.notconnected",   nil}
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

// Param adds a param
func (status Status) Param(key string, value interface{}) (Status) {
	final := status
	if len(final.Params) == 0 {
		final.Params = map[string]interface{}{}
	}
	final.Params[key] = value
	return final
}

func (status Status) Error() string {
	if status.IsOK() {
		return ""
	}
	if len(status.Params )> 0 {
		sb := strings.Builder{}
		sb.WriteString(status.Reason)
		sb.WriteString("(")
		first := true
		for key, value := range status.Params {
			if first {
				first = false
			} else {
				sb.WriteString(", ")
			}
			sb.WriteString(key)
			sb.WriteString(": ")
			sb.WriteString(fmt.Sprintf("%v", value))
		}
		sb.WriteString(")")
		return sb.String()
	}
	return status.Reason
}