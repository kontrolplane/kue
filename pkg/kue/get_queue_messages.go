package kue

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

func FetchQueueMessages(client *sqs.Client, ctx context.Context, queueUrl string, maxMessages int32) ([]Message, error) {

	input := &sqs.ReceiveMessageInput{
		QueueUrl:              &queueUrl,
		MaxNumberOfMessages:   maxMessages,
		VisibilityTimeout:     1, // Minimal visibility timeout to keep messages visible
		MessageAttributeNames: []string{"All"},
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
	}

	result, err := client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	messages := make([]Message, 0, len(result.Messages))
	for _, msg := range result.Messages {
		message := Message{
			MessageID:     *msg.MessageId,
			Body:          *msg.Body,
			MD5OfBody:     *msg.MD5OfBody,
			ReceiptHandle: *msg.ReceiptHandle,
		}

		// Handle attributes
		if msg.Attributes != nil {
			message.Attributes = make(map[string]string)
			for key, value := range msg.Attributes {
				message.Attributes[key] = value
			}

			// Parse specific attributes
			if sentTimestamp, ok := msg.Attributes[string(types.MessageSystemAttributeNameSentTimestamp)]; ok {
				if timestamp, err := strconv.ParseInt(sentTimestamp, 10, 64); err == nil {
					message.SentTimestamp = time.Unix(timestamp/1000, 0).Format(time.RFC3339)
				}
			}

			if approximateFirstReceiveTimestamp, ok := msg.Attributes[string(types.MessageSystemAttributeNameApproximateFirstReceiveTimestamp)]; ok {
				if timestamp, err := strconv.ParseInt(approximateFirstReceiveTimestamp, 10, 64); err == nil {
					message.FirstReceiveTime = time.Unix(timestamp/1000, 0).Format(time.RFC3339)
				}
			}

			if receiveCount, ok := msg.Attributes[string(types.MessageSystemAttributeNameApproximateReceiveCount)]; ok {
				message.ReceiveCount = receiveCount
			}

			// Handle FIFO queue specific attributes
			if messageGroupId, ok := msg.Attributes[string(types.MessageSystemAttributeNameMessageGroupId)]; ok {
				message.MessageGroupID = messageGroupId
			}

			if messageDeduplicationId, ok := msg.Attributes[string(types.MessageSystemAttributeNameMessageDeduplicationId)]; ok {
				message.MessageDeduplicationID = messageDeduplicationId
			}

			if sequenceNumber, ok := msg.Attributes[string(types.MessageSystemAttributeNameSequenceNumber)]; ok {
				message.SequenceNumber = sequenceNumber
			}
		}

		// Handle message attributes
		if msg.MessageAttributes != nil {
			message.MessageAttributes = make(map[string]string)
			for key, attr := range msg.MessageAttributes {
				if attr.StringValue != nil {
					message.MessageAttributes[key] = *attr.StringValue
				}
			}
		}

		messages = append(messages, message)
	}

	return messages, nil
}


