package kue

import (
    "context"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// DeleteMessage deletes a specific SQS message given the queue URL and the
// receipt handle returned when the message was received.
//
// The AWS SDK only allows deleting messages using the receipt handle as a
// proof-of-work that we have processed (or at least received) the message.
func DeleteMessage(client *sqs.Client, ctx context.Context, queueUrl string, receiptHandle string) error {
    _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
        QueueUrl:      &queueUrl,
        ReceiptHandle: &receiptHandle,
    })
    return err
}
