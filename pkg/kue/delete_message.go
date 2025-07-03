package kue

import (
    "context"
    "log"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// DeleteMessage deletes a single message from the queue given the receipt handle.
//
// queueUrl must be the full URL of the queue; receiptHandle is obtained from
// a prior ReceiveMessage call.
func DeleteMessage(client *sqs.Client, ctx context.Context, queueUrl string, receiptHandle string) error {
    _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
        QueueUrl:      &queueUrl,
        ReceiptHandle: &receiptHandle,
    })
    if err != nil {
        return err
    }
    log.Printf("[DeleteMessage] Deleted message with receipt handle: %s", receiptHandle)
    return nil
}
