package kue

import (
    "context"
    "fmt"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueue creates a new SQS queue with the provided name. If fifo is true a FIFO queue
// will be created, otherwise a standard queue will be created. Only a subset of queue
// attributes are currently supported; more can be added when we extend the create form.
//
// The function returns an error if the AWS SDK call fails or if mandatory parameters are
// missing/invalid.
func CreateQueue(client *sqs.Client, ctx context.Context, name string, fifo bool) error {
    if name == "" {
        return fmt.Errorf("queue name cannot be empty")
    }

    // Ensure the .fifo suffix when requested
    if fifo && !strings.HasSuffix(name, ".fifo") {
        name = name + ".fifo"
    }

    attrs := make(map[string]string)
    if fifo {
        attrs["FifoQueue"] = "true"
        // Enable content-based deduplication by default – this is the most common
        attrs["ContentBasedDeduplication"] = "true"
    }

    _, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
        QueueName:  aws.String(name),
        Attributes: attrs,
    })
    return err
}
