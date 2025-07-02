package kue

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueueInput represents the subset of SQS queue attributes we currently expose in the TUI.
// Additional optional attributes can be added over time without breaking callers because
// they will retain their zero-value which we interpret as "use AWS default".
type CreateQueueInput struct {
    // Name is the queue name that will be visible in the AWS console.
    Name string

    // TODO: surface these through the UI in follow-up patches
    VisibilityTimeout      int32
    MessageRetentionPeriod int32
    DelaySeconds           int32
    ReceiveMessageWaitTime int32
    MaximumMessageSize     int32
}

// CreateQueue wraps the AWS SDK CreateQueue call translating only the non-zero
// fields into the AWS Attributes map so that AWS continues to apply its defaults
// for every attribute the user did not explicitly set.
func CreateQueue(client *sqs.Client, ctx context.Context, input CreateQueueInput) (string, error) {
    if input.Name == "" {
        return "", fmt.Errorf("queue name is required")
    }

    attrs := make(map[string]string)

    if input.VisibilityTimeout != 0 {
        attrs["VisibilityTimeout"] = intToString(input.VisibilityTimeout)
    }
    if input.MessageRetentionPeriod != 0 {
        attrs["MessageRetentionPeriod"] = intToString(input.MessageRetentionPeriod)
    }
    if input.DelaySeconds != 0 {
        attrs["DelaySeconds"] = intToString(input.DelaySeconds)
    }
    if input.ReceiveMessageWaitTime != 0 {
        attrs["ReceiveMessageWaitTimeSeconds"] = intToString(input.ReceiveMessageWaitTime)
    }
    if input.MaximumMessageSize != 0 {
        attrs["MaximumMessageSize"] = intToString(input.MaximumMessageSize)
    }

    out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
        QueueName:  aws.String(input.Name),
        Attributes: attrs,
    })
    if err != nil {
        return "", err
    }

    return aws.ToString(out.QueueUrl), nil
}

func intToString(v int32) string {
    return fmt.Sprintf("%d", v)
}

