package kue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// PurgeQueue purges all messages from the queue at the given URL.
func PurgeQueue(client *sqs.Client, ctx context.Context, queueUrl string) error {
	_, err := client.PurgeQueue(ctx, &sqs.PurgeQueueInput{
		QueueUrl: &queueUrl,
	})
	if err != nil {
		return err
	}

	log.Printf("[PurgeQueue] Purged queue: %s", queueUrl)
	return nil
}
