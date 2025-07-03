package kue

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    sqsTypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// redrivePolicy is a helper struct to unmarshal the RedrivePolicy attribute
// returned by GetQueueAttributes.
type redrivePolicy struct {
    DeadLetterTargetArn string `json:"deadLetterTargetArn"`
    MaxReceiveCount      string `json:"maxReceiveCount"`
}

// DeleteMessage deletes a message from the specified queue by its receipt handle.
func DeleteMessage(ctx context.Context, client *sqs.Client, queueURL string, receiptHandle string) error {
    _, err := client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
        QueueUrl:      aws.String(queueURL),
        ReceiptHandle: aws.String(receiptHandle),
    })
    if err != nil {
        return fmt.Errorf("failed to delete message from queue %s: %w", queueURL, err)
    }
    return nil
}

// MoveMessageToDLQ moves a message to the queue's configured DLQ and then deletes it from the source queue.
//
// The function performs the following steps:
//   1. Fetches the RedrivePolicy attribute of the source queue to determine the DLQ ARN.
//   2. Resolves the DLQ URL via GetQueueUrl.
//   3. Sends the message body and attributes to the DLQ using SendMessage.
//   4. Upon successful send, deletes the original message from the source queue.
func MoveMessageToDLQ(ctx context.Context, client *sqs.Client, queueURL string, message Message) error {
    // Step 1: obtain RedrivePolicy for provided queue
    attrOut, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
        QueueUrl:       &queueURL,
        AttributeNames: []sqsTypes.QueueAttributeName{sqsTypes.QueueAttributeNameRedrivePolicy},
    })
    if err != nil {
        return fmt.Errorf("failed to get queue attributes for %s: %w", queueURL, err)
    }

    redriveStr, ok := attrOut.Attributes[string(sqsTypes.QueueAttributeNameRedrivePolicy)]
    if !ok || strings.TrimSpace(redriveStr) == "" {
        return fmt.Errorf("queue %s has no RedrivePolicy configured; cannot move message to DLQ", queueURL)
    }

    var rp redrivePolicy
    if err := json.Unmarshal([]byte(redriveStr), &rp); err != nil {
        return fmt.Errorf("unable to parse RedrivePolicy for queue %s: %w", queueURL, err)
    }

    if rp.DeadLetterTargetArn == "" {
        return fmt.Errorf("RedrivePolicy for queue %s does not contain deadLetterTargetArn", queueURL)
    }

    // Extract queue name from ARN (last segment after ':')
    arnParts := strings.Split(rp.DeadLetterTargetArn, ":")
    if len(arnParts) < 6 {
        return fmt.Errorf("invalid DLQ ARN %s", rp.DeadLetterTargetArn)
    }
    dlqName := arnParts[len(arnParts)-1]

    // Step 2: get DLQ URL
    getUrlOut, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{QueueName: &dlqName})
    if err != nil {
        return fmt.Errorf("failed to get URL for DLQ %s: %w", dlqName, err)
    }
    dlqURL := *getUrlOut.QueueUrl

    // Step 3: send message to DLQ preserving attributes where possible
    sendInput := &sqs.SendMessageInput{
        QueueUrl:       &dlqURL,
        MessageBody:    &message.Body,
    }

    // convert custom map[string]string to SQS MessageAttributeValue map if present
    if len(message.MessageAttributes) > 0 {
        attrs := make(map[string]sqsTypes.MessageAttributeValue)
        for k, v := range message.MessageAttributes {
            val := v
            attrs[k] = sqsTypes.MessageAttributeValue{
                DataType:    aws.String("String"),
                StringValue: aws.String(val),
            }
        }
        sendInput.MessageAttributes = attrs
    }

    // Preserve FIFO-specific attributes if present in message attributes
    if message.MessageGroupID != "" {
        sendInput.MessageGroupId = aws.String(message.MessageGroupID)
    }
    if message.MessageDeduplicationID != "" {
        sendInput.MessageDeduplicationId = aws.String(message.MessageDeduplicationID)
    }

    if _, err := client.SendMessage(ctx, sendInput); err != nil {
        return fmt.Errorf("failed to send message to DLQ %s: %w", dlqURL, err)
    }

    // Step 4: delete original message
    if err := DeleteMessage(ctx, client, queueURL, message.ReceiptHandle); err != nil {
        return fmt.Errorf("message was sent to DLQ %s but failed to delete from source queue %s: %w", dlqURL, queueURL, err)
    }

    return nil
}
