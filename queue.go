package iwt

// Queue describe a queue
type Queue struct {
	Name               string    `json:"queueName"`
	Type               QueueType `json:"queueType"`
	EstimatedWaitTime  int       `json:"estimatedWaitTime"`  // in H
	PollWaitSuggestion int       `json:"pollWaitSuggestion"` // in ms
	AvailableAgents    int       `json:"agentsAvailable"`
	Status             Status    `json:"status"`
}

// QueryQueue queries a queue for its status
func (client *Client) QueryQueue(queuename string, queuetype QueueType) (*Queue, error) {
	results := struct{Queue Queue `json:"queue"`}{}
	_, _, err := client.sendRequest(client.Context, &requestOptions{
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
	return queue.Type.String() + " Queue:" + queue.Name
}