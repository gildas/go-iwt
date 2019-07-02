package iwt

import (
	"strings"
)

// Queue describe a queue
type Queue struct {
	Name               string    `json:"queueName"`
	Type               QueueType `json:"queueType"`
	EstimatedWaitTime  int       `json:"estimatedWaitTime"`  // in H
	PollWaitSuggestion int       `json:"pollWaitSuggestion"` // in ms
	AvailableAgents    int       `json:"agentsAvailable"`
	Status             Status    `json:"status"`
}

// NewQueue instantiates a new Queue
// The Qualified Queue name follows the PureConnect standard way of specifying a queue:
// "User Queue:Operator", "Workgroup Queue:Sales", "Station Queue:7001"
// Workgroup Queue is the default type, and if nothing was given, "Workgroup Queue:CompanyOperator" will be used
func NewQueue(qualifiedqueue string) *Queue {
	queueparts := strings.Split(strings.TrimSpace(qualifiedqueue), ":")
	if len(queueparts) > 1 {
		queuename := strings.Join(queueparts[1:], ":")
		switch strings.ToLower(strings.TrimSpace(queueparts[0])) {
		case "station queue", "stationqueue", "station":
			return &Queue{Name: queuename, Type: StationQueue}
		case "user queue", "userqueue", "user":
			return &Queue{Name: queuename, Type: UserQueue}
		case "workgroup queue", "workgroupqueue", "workgroup":
			return &Queue{Name: queuename, Type: WorkgroupQueue}
		default:
			return &Queue{Name: queuename, Type: WorkgroupQueue}
		}
	} else if len(queueparts) > 0 {
		return &Queue{Name: queueparts[0], Type: WorkgroupQueue}
	} else {
		return &Queue{Name: "CompanyOperator", Type: WorkgroupQueue}
	}
}

// QueryQueue queries a queue for its status
func (client *Client) QueryQueue(queuename string, queuetype QueueType) (*Queue, error) {
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

func (queue *Queue) String() string {
	return queue.Type.Prefix() + queue.Name
}