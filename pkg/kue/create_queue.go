package kue

import (
    "context"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// QueueCreateInput represents the parameters used to create a queue.
type QueueCreateInput struct {
    // Name of the queue (required)
    Name string

    // Attributes is an optional map of additional queue attributes.
    Attributes map[string]string
}

// CreateQueue creates a new SQS queue and returns its url.
func CreateQueue(client *sqs.Client, ctx context.Context, input QueueCreateInput) (string, error) {

    if input.Attributes == nil {
        input.Attributes = map[string]string{}
    }

    output, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
        QueueName:  aws.String(input.Name),
        Attributes: input.Attributes,
    })
    if err != nil {
        return "", err
    }

    log.Printf("[CreateQueue] Created queue %s at url %s", input.Name, aws.ToString(output.QueueUrl))

    return aws.ToString(output.QueueUrl), nil
}
