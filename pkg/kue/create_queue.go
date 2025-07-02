package kue

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueue creates a new SQS queue using the provided attributes.
// The attributes map expects keys that match the SQS attribute names, e.g.
//   - "VisibilityTimeout"
//   - "MessageRetentionPeriod"
//   - "DelaySeconds"
//   - "MaximumMessageSize"
//   - "ReceiveMessageWaitTimeSeconds"
// If the attributes parameter is nil an empty attribute map will be passed along.
//
// The function returns the queue url when the operation succeeds or an error otherwise.
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
