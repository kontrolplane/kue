package kue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// SendMessageInput contains the parameters for sending a message.
type SendMessageInput struct {
	QueueUrl    string
	MessageBody string
	// For FIFO queues
	MessageGroupId         string
	MessageDeduplicationId string
}

// SendMessage sends a message to an SQS queue.
func SendMessage(client *sqs.Client, ctx context.Context, input SendMessageInput) error {
	sqsInput := &sqs.SendMessageInput{
		QueueUrl:    &input.QueueUrl,
		MessageBody: &input.MessageBody,
	}

	// Add FIFO queue attributes if provided
	if input.MessageGroupId != "" {
		sqsInput.MessageGroupId = &input.MessageGroupId
	}
	if input.MessageDeduplicationId != "" {
		sqsInput.MessageDeduplicationId = &input.MessageDeduplicationId
	}

	_, err := client.SendMessage(ctx, sqsInput)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
