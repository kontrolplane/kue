package kue

import (
    "context"
    "fmt"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueue creates a new SQS queue with the given name and attribute map. If
// the queue already exists, the call is idempotent and AWS will return the
// existing URL.
//
// attrs must use the exact attribute keys as documented in the AWS API (e.g.
// "VisibilityTimeout", "MessageRetentionPeriod"). All values must be strings.
//
// The function returns the created queue URL or an error.
func CreateQueue(client *sqs.Client, ctx context.Context, name string, attrs map[string]string) (string, error) {

    if name == "" {
        return "", fmt.Errorf("queue name must not be empty")
    }

    input := &sqs.CreateQueueInput{
        QueueName:  aws.String(name),
        Attributes: attrs,
    }

    out, err := client.CreateQueue(ctx, input)
    if err != nil {
        return "", err
    }

    log.Printf("[CreateQueue] Created queue %s (%s)", name, aws.ToString(out.QueueUrl))
    return aws.ToString(out.QueueUrl), nil
}

