package kue

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// ChangeMessageVisibility sets the visibility timeout for a specific message identified by its receipt handle.
// A visibilityTimeout of 0 will immediately make the message visible again (i.e., re-queue it).
func ChangeMessageVisibility(client *sqs.Client, ctx context.Context, queueUrl string, receiptHandle string, visibilityTimeout int32) error {
    input := &sqs.ChangeMessageVisibilityInput{
        QueueUrl:          &queueUrl,
        ReceiptHandle:     &receiptHandle,
        VisibilityTimeout: visibilityTimeout,
    }

    if _, err := client.ChangeMessageVisibility(ctx, input); err != nil {
        return fmt.Errorf("failed to change message visibility: %w", err)
    }

    return nil
}
