package kue

import (
    "context"
    "log"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// DeleteMessage deletes a specific SQS message identified by its receipt handle
func DeleteMessage(client *sqs.Client, ctx context.Context, queueUrl string, receiptHandle string) error {
    _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
        QueueUrl:      aws.String(queueUrl),
        ReceiptHandle: aws.String(receiptHandle),
    })
    if err != nil {
        return err
    }

    log.Printf("[DeleteMessage] Deleted message with handle: %s", receiptHandle)
    return nil
}
