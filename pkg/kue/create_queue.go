package kue

import (
    "context"
    "log"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueue creates a new SQS queue with the provided name and optional attributes.
// The attributes map should use the exact attribute keys expected by the SQS CreateQueue API
// (e.g. DelaySeconds, VisibilityTimeout, MessageRetentionPeriod, etc.).
//
// If attributes is nil or empty, the queue will be created with AWS defaults.
func CreateQueue(client *sqs.Client, ctx context.Context, queueName string, attributes map[string]string) error {
    _, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
        QueueName:  &queueName,
        Attributes: attributes,
    })
    if err != nil {
        return err
    }

    log.Printf("[CreateQueue] Created queue: %s", queueName)
    return nil
}

