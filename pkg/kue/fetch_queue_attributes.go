package kue

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// FetchQueueAttributes fetches the details of a queue
func FetchQueueAttributes(client *sqs.Client, ctx context.Context, queueUrl string) (Queue, error) {

	attrsResult, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       &queueUrl,
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		return Queue{}, fmt.Errorf("failed to get queue attributes: %w", err)
	}

	urlParts := strings.Split(queueUrl, "/")
	queue := Queue{
		Url:  queueUrl,
		Name: urlParts[len(urlParts)-1],
	}

	attributeMap := map[types.QueueAttributeName]*string{
		types.QueueAttributeNameCreatedTimestamp:                      &queue.CreatedTimestamp,
		types.QueueAttributeNameLastModifiedTimestamp:                 &queue.LastModified,
		types.QueueAttributeNameDelaySeconds:                          &queue.DelaySeconds,
		types.QueueAttributeNameMaximumMessageSize:                    &queue.MaxMessageSize,
		types.QueueAttributeNameMessageRetentionPeriod:                &queue.MessageRetentionPeriod,
		types.QueueAttributeNameReceiveMessageWaitTimeSeconds:         &queue.ReceiveMessageWaitTime,
		types.QueueAttributeNameVisibilityTimeout:                     &queue.VisibilityTimeout,
		types.QueueAttributeNameFifoQueue:                             &queue.IsFifo,
		types.QueueAttributeNameContentBasedDeduplication:             &queue.ContentBasedDeduplication,
		types.QueueAttributeNameApproximateNumberOfMessages:           &queue.ApproximateNumberOfMessages,
		types.QueueAttributeNameApproximateNumberOfMessagesNotVisible: &queue.ApproximateNumberOfMessagesNotVisible,
		types.QueueAttributeNameApproximateNumberOfMessagesDelayed:    &queue.ApproximateNumberOfMessagesDelayed,
	}

	for attrName, field := range attributeMap {
		if val, ok := attrsResult.Attributes[string(attrName)]; ok {
			*field = val
		}
	}

	if tagsResult, err := client.ListQueueTags(ctx, &sqs.ListQueueTagsInput{QueueUrl: &queueUrl}); err == nil && len(tagsResult.Tags) > 0 {
		queue.Tags = tagsResult.Tags
	}

	return queue, nil
}
