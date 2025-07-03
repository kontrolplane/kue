package kue

import (
    "context"
    "strconv"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// CreateQueueAttributes represents optional settings accepted by CreateQueue
// All zero values will be ignored and not sent to SQS
// Units are seconds unless otherwise specified by AWS
// Reference: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_CreateQueue.html

type CreateQueueAttributes struct {
    VisibilityTimeout      int  // seconds
    MessageRetentionPeriod int  // seconds
    DelaySeconds           int  // seconds
    MaximumMessageSize     int  // bytes
    ReceiveMessageWaitTime int  // seconds
    FifoQueue              bool // only considered when queue name ends with .fifo
}

// CreateQueue creates a queue with the given name and optional attributes
// If queue already exists, it will return an error from AWS
func CreateQueue(client *sqs.Client, ctx context.Context, name string, attrs CreateQueueAttributes) error {
    a := make(map[string]string)

    if attrs.VisibilityTimeout != 0 {
        a[string(types.QueueAttributeNameVisibilityTimeout)] = intToString(attrs.VisibilityTimeout)
    }
    if attrs.MessageRetentionPeriod != 0 {
        a[string(types.QueueAttributeNameMessageRetentionPeriod)] = intToString(attrs.MessageRetentionPeriod)
    }
    if attrs.DelaySeconds != 0 {
        a[string(types.QueueAttributeNameDelaySeconds)] = intToString(attrs.DelaySeconds)
    }
    if attrs.MaximumMessageSize != 0 {
        a[string(types.QueueAttributeNameMaximumMessageSize)] = intToString(attrs.MaximumMessageSize)
    }
    if attrs.ReceiveMessageWaitTime != 0 {
        a[string(types.QueueAttributeNameReceiveMessageWaitTimeSeconds)] = intToString(attrs.ReceiveMessageWaitTime)
    }
    if attrs.FifoQueue {
        a[string(types.QueueAttributeNameFifoQueue)] = "true"
    }

    _, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
        QueueName:  aws.String(name),
        Attributes: a,
    })
    return err
}

func intToString(v int) string {
    return strconv.Itoa(v)
}
