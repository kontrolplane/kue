package kue

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueue creates a new SQS queue with optional attributes and returns the queue URL.
// Provide a map of attributes where the keys correspond to SQS attribute names such as
//   VisibilityTimeout, MessageRetentionPeriod, DelaySeconds, MaximumMessageSize, ReceiveMessageWaitTimeSeconds.
// Pass nil if no attributes need to be set.
func CreateQueue(client *sqs.Client, ctx context.Context, queueName string, attributes map[string]string) (string, error) {
    if attributes == nil {
        attributes = map[string]string{}
    }

    out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
        QueueName:  aws.String(queueName),
        Attributes: attributes,
    })
    if err != nil {
        return "", err
    }

    return aws.ToString(out.QueueUrl), nil
}

