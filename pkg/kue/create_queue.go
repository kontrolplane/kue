package kue

import (
    "context"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// CreateQueue creates a new SQS queue using the supplied attributes and returns a populated Queue struct.
// The attributes map must follow the AWS SQS CreateQueue attribute keys, e.g. "DelaySeconds", "FifoQueue", etc.
func CreateQueue(client *sqs.Client, ctx context.Context, name string, attributes map[string]string, tags map[string]string) (Queue, error) {

    input := &sqs.CreateQueueInput{
        QueueName:  aws.String(name),
        Attributes: attributes,
    }

    if len(tags) > 0 {
        input.Tags = tags
    }

    // Amazon SQS requires the ContentBasedDeduplication parameter only for FIFO queues.
    // This logic is handled by the caller via attributes.

    out, err := client.CreateQueue(ctx, input)
    if err != nil {
        return Queue{}, err
    }

    // Fetch attributes immediately to return a fully-populated Queue struct.
    q, err := FetchQueueAttributes(client, ctx, aws.ToString(out.QueueUrl))
    if err != nil {
        // Fall back to returning minimal Queue info if attribute fetching fails.
        parts := strings.Split(aws.ToString(out.QueueUrl), "/")
        return Queue{Url: aws.ToString(out.QueueUrl), Name: parts[len(parts)-1]}, nil
    }

    return q, nil
}
