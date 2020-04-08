package iwt_test

import (
	"encoding/json"
	"testing"

	"github.com/gildas/go-iwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueTypeCanStringify(t *testing.T) {
	queueType := iwt.WorkgroupQueue
	assert.Equal(t, "Workgroup", queueType.String())
}

func TestQueueTypeCanMarshal(t *testing.T) {
	queue := struct {
		QueueName string        `json:"queueName"`
		QueueType iwt.QueueType `json:"queueType"`
	}{"test", iwt.WorkgroupQueue}

	data, err := json.Marshal(queue)
	require.Nil(t, err, "Cannot marshal, error: %s", err)
	assert.JSONEq(t, string(data), `{"queueName":"test","queueType":"Workgroup"}`)
}

func TestQueueTypeCanUnmarshal(t *testing.T) {
	payload := []byte(`{"queueName":"test","queueType":"Workgroup"}`)
	queue := struct {
		QueueName string        `json:"queueName"`
		QueueType iwt.QueueType `json:"queueType"`
	}{}

	err := json.Unmarshal(payload, &queue)
	require.Nil(t, err, "Cannot unmarshal, error: %s", err)
	assert.Equal(t, "test", queue.QueueName)
	assert.Equal(t, iwt.WorkgroupQueue, queue.QueueType)
}
