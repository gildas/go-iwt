package iwt

import (
	"fmt"
)

// Queue describe a queue
type Queue struct {
	Name               string `json:"queueName"`
	Type               string `json:"queueType"`
	EstimatedWaitTime  int    `json:"estimatedWaitTime"`  // in H
	PollWaitSuggestion int    `json:"pollWaitSuggestion"` // in ms
	AvailableAgents    int    `json:"agentsAvailable"`
	Status             Status `json:"status"`
}

// QueryQueue queries a queue for its status
func (client *Client) QueryQueue(queuename, queuetype string) (*Queue, error) {
	switch {
	case len(queuetype) == 0:
		queuetype = "Workgroup"
	case queuetype == "Station":
	case queuetype == "User":
	case queuetype == "Workgroup":
	default:
		return nil, fmt.Errorf("error.websvc.unknownEntity.invalidQueueType")
	}
	results := struct{Queue Queue `json:"queue"`}{}
	_, err := client.sendRequest(client.Context, &requestOptions{
		Path: "/queue/query",
		Payload: struct {
			Queue
			Participant Participant `json:"participant"`
		}{
			Queue{Name: queuename, Type: queuetype},
			Participant{Name: "Anonymous User"},
		},
	}, &results)
	if err != nil {
		return nil, err
	}
	results.Queue.Name = queuename
	results.Queue.Type = queuetype
	return &results.Queue, results.Queue.Status.AsError()
}