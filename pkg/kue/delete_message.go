package kue

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// DeleteMessage deletes a single message from the given queue url using the provided receipt handle.
// An error is returned when the AWS SDK operation fails.
func DeleteMessage(client *sqs.Client, ctx context.Context, queueUrl string, receiptHandle string) error {
    input := &sqs.DeleteMessageInput{
        QueueUrl:      &queueUrl,
        ReceiptHandle: &receiptHandle,
    }

    if _, err := client.DeleteMessage(ctx, input); err != nil {
        return fmt.Errorf("failed to delete message: %w", err)
    }

    return nil
}
