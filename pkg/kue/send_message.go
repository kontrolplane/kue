package kue

import (
    "context"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SendMessage sends a message to the provided SQS queue URL and returns the message ID.
func SendMessage(client *sqs.Client, ctx context.Context, queueUrl string, body string, delaySeconds int32) (string, error) {
    input := &sqs.SendMessageInput{
        QueueUrl:    &queueUrl,
        MessageBody: &body,
    }
    if delaySeconds > 0 {
        input.DelaySeconds = delaySeconds
    }

    result, err := client.SendMessage(ctx, input)
    if err != nil {
        return "", fmt.Errorf("failed to send message: %w", err)
    }
    if result.MessageId == nil {
        return "", fmt.Errorf("message ID nil after send")
    }

    return *result.MessageId, nil
}
