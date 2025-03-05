package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// ListQueuesByPrefix lists the SQS queues by a given prefix
func ListQueuesByPrefix(client *sqs.Client, ctx context.Context) (queues []Queue, err error) {
	return nil, nil
}
