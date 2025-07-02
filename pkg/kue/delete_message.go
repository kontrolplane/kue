package kue

import (
    "context"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// DeleteMessage removes a single message from the specified queue using its receipt handle.
// The receipt handle is required by SQS to uniquely identify the message to delete.
func DeleteMessage(client *sqs.Client, ctx context.Context, queueUrl string, receiptHandle string) error {
    _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
        QueueUrl:      aws.String(queueUrl),
        ReceiptHandle: aws.String(receiptHandle),
    })
    if err != nil {
        return err
    }

    log.Printf("[DeleteMessage] Deleted message from queue %s", queueUrl)
    return nil
}
