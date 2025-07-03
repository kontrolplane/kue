package kue

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SQSMessageDeleter defines the subset of AWS SDK client functions we need for deleting messages.
// *sqs.Client already implements this interface, which makes the helper functions easy to test with
// stub implementations.
type SQSMessageDeleter interface {
    DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

// DeleteMessage deletes a single message from the queue identified by queueURL using the
// provided receiptHandle.
func DeleteMessage(client SQSMessageDeleter, ctx context.Context, queueURL, receiptHandle string) error {
    if client == nil {
        return fmt.Errorf("nil sqs client")
    }
    if queueURL == "" {
        return fmt.Errorf("queueURL cannot be empty")
    }
    if receiptHandle == "" {
        return fmt.Errorf("receiptHandle cannot be empty")
    }

    _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
        QueueUrl:      &queueURL,
        ReceiptHandle: &receiptHandle,
    })
    if err != nil {
        return fmt.Errorf("delete message failed: %w", err)
    }

    return nil
}

// DeleteMessages deletes multiple messages. It returns a slice of errors matching the length of
// receiptHandles. A nil entry indicates success for that index.
func DeleteMessages(client SQSMessageDeleter, ctx context.Context, queueURL string, receiptHandles []string) []error {
    errs := make([]error, len(receiptHandles))
    for i, rh := range receiptHandles {
        errs[i] = DeleteMessage(client, ctx, queueURL, rh)
    }
    return errs
}