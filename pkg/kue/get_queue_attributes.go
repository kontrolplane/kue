package kue

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
		types.QueueAttributeNameQueueArn:                              &queue.Arn,
		types.QueueAttributeNameDelaySeconds:                          &queue.DelaySeconds,
		types.QueueAttributeNameMaximumMessageSize:                    &queue.MaxMessageSize,
		types.QueueAttributeNameMessageRetentionPeriod:                &queue.MessageRetentionPeriod,
		types.QueueAttributeNameReceiveMessageWaitTimeSeconds:         &queue.ReceiveMessageWaitTime,
		types.QueueAttributeNameVisibilityTimeout:                     &queue.VisibilityTimeout,
		types.QueueAttributeNameContentBasedDeduplication:             &queue.ContentBasedDeduplication,
		types.QueueAttributeNameApproximateNumberOfMessages:           &queue.ApproximateNumberOfMessages,
		types.QueueAttributeNameApproximateNumberOfMessagesNotVisible: &queue.ApproximateNumberOfMessagesNotVisible,
		types.QueueAttributeNameApproximateNumberOfMessagesDelayed:    &queue.ApproximateNumberOfMessagesDelayed,
	}

	if val, ok := attrsResult.Attributes[string(types.QueueAttributeNameCreatedTimestamp)]; ok {
		createdTime, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			queue.CreatedTimestamp = time.Unix(createdTime, 0).Format("2006-01-02 15:04:05")
		}
	}

	if val, ok := attrsResult.Attributes[string(types.QueueAttributeNameLastModifiedTimestamp)]; ok {
		modifiedTime, err := strconv.ParseInt(val, 10, 64)
		if err == nil {
			queue.LastModified = time.Unix(modifiedTime, 0).Format("2006-01-02 15:04:05")
		}
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
