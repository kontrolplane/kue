package kue

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/kontrolplane/kue/pkg/audit"
)

// DeleteMessage removes a single message from the given SQS queue. It uses the message's receipt
// handle as required by the AWS API. On successful deletion, it appends an audit log entry that
// captures the queue name, message ID, timestamp, and user performing the action.
//
// The audit write happens on a best-effort basis – if it fails, the error is returned but the
// deletion has already happened. Callers may choose to ignore the audit error if they do not wish
// to block user flows.
func DeleteMessage(client *sqs.Client, ctx context.Context, queueURL, queueName, receiptHandle, messageID string) error {
    _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
        QueueUrl:      &queueURL,
        ReceiptHandle: &receiptHandle,
    })
    if err != nil {
        return fmt.Errorf("failed to delete message: %w", err)
    }

    // Attempt to record audit; return error if it fails but cannot roll back deletion.
    if auditErr := audit.AppendDeletion(queueName, messageID); auditErr != nil {
        // Wrap but keep original error semantics.
        return fmt.Errorf("message deleted but failed to write audit log: %w", auditErr)
    }

    return nil
}
