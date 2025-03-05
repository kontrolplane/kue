package sqs

import (
	"fmt"
	"strings"
)

// format extracts the queue struct from a queue URL
func formatQueueName(queueUrl string) (Queue, error) {
	parts := strings.Split(queueUrl, "/")
	if len(parts) < 5 {
		return Queue{}, fmt.Errorf("Invalid queue URL: %s", queueUrl)
	}

	return Queue{
		Protocol:          strings.TrimSuffix(parts[0], ":"),
		ServiceEndpoint:   parts[2],
		AccountIdentifier: parts[3],
		Name:              parts[4],
	}, nil
}
