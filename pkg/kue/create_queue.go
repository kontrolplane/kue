package kue

import (
	"context"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// QueueConfig holds the configuration for creating a new SQS queue
type QueueConfig struct {
	Name                      string
	IsFifo                    bool
	DelaySeconds              int    // 0-900 seconds (default: 0)
	MaximumMessageSize        int    // 1024-262144 bytes (default: 262144 = 256KB)
	MessageRetentionPeriod    int    // 60-1209600 seconds (default: 345600 = 4 days)
	ReceiveMessageWaitTime    int    // 0-20 seconds (default: 0)
	VisibilityTimeout         int    // 0-43200 seconds (default: 30)
	ContentBasedDeduplication bool   // FIFO only
	DeduplicationScope        string // FIFO only: "messageGroup" or "queue"
	FifoThroughputLimit       string // FIFO only: "perQueue" or "perMessageGroupId"
}

// CreateQueue creates a new SQS queue with the provided configuration
func CreateQueue(client *sqs.Client, ctx context.Context, config QueueConfig) (*string, error) {
	queueName := config.Name
	if config.IsFifo && len(queueName) > 0 {
		if len(queueName) < 5 || queueName[len(queueName)-5:] != ".fifo" {
			queueName = queueName + ".fifo"
		}
	}

	attributes := make(map[string]string)

	// Standard queue attributes
	if config.DelaySeconds > 0 {
		attributes[string(types.QueueAttributeNameDelaySeconds)] = strconv.Itoa(config.DelaySeconds)
	}
	if config.MaximumMessageSize > 0 {
		attributes[string(types.QueueAttributeNameMaximumMessageSize)] = strconv.Itoa(config.MaximumMessageSize)
	}
	if config.MessageRetentionPeriod > 0 {
		attributes[string(types.QueueAttributeNameMessageRetentionPeriod)] = strconv.Itoa(config.MessageRetentionPeriod)
	}
	if config.ReceiveMessageWaitTime > 0 {
		attributes[string(types.QueueAttributeNameReceiveMessageWaitTimeSeconds)] = strconv.Itoa(config.ReceiveMessageWaitTime)
	}
	if config.VisibilityTimeout > 0 {
		attributes[string(types.QueueAttributeNameVisibilityTimeout)] = strconv.Itoa(config.VisibilityTimeout)
	}

	// FIFO queue attributes
	if config.IsFifo {
		attributes[string(types.QueueAttributeNameFifoQueue)] = "true"
		if config.ContentBasedDeduplication {
			attributes[string(types.QueueAttributeNameContentBasedDeduplication)] = "true"
		}
		if config.DeduplicationScope != "" {
			attributes[string(types.QueueAttributeNameDeduplicationScope)] = config.DeduplicationScope
		}
		if config.FifoThroughputLimit != "" {
			attributes[string(types.QueueAttributeNameFifoThroughputLimit)] = config.FifoThroughputLimit
		}
	}

	input := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	}

	if len(attributes) > 0 {
		input.Attributes = attributes
	}

	result, err := client.CreateQueue(ctx, input)
	if err != nil {
		log.Printf("[CreateQueue] Error creating queue %s: %v", queueName, err)
		return nil, err
	}

	log.Printf("[CreateQueue] Created queue: %s, URL: %s", queueName, *result.QueueUrl)

	return result.QueueUrl, nil
}
