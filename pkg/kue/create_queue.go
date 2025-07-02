package kue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueue creates a new SQS queue with the provided name and optional attributes.
// The attributes map follows the keys as documented by AWS, e.g. VisibilityTimeout,
// MessageRetentionPeriod, DelaySeconds, MaximumMessageSize, ReceiveMessageWaitTimeSeconds.
// Any attribute whose value is an empty string will be ignored.
func CreateQueue(client *sqs.Client, ctx context.Context, queueName string, attributes map[string]string) error {
	if queueName == "" {
		return fmt.Errorf("queue name cannot be empty")
	}

	attribs := make(map[string]string)
	for k, v := range attributes {
		if v != "" {
			attribs[k] = v
		}
	}

	_, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName:  &queueName,
		Attributes: attribs,
	})
	if err != nil {
		return fmt.Errorf("error creating queue: %w", err)
	}

	return nil
}

