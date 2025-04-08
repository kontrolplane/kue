package kue

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// DeleteQueue deletes the queue for the queueUrl passed
func DeleteQueue(client *sqs.Client, ctx context.Context, queueName string) error {

	urlResult, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
		QueueName: aws.String(queueName),
	})
	if err != nil {
		return err
	}

	_, err = client.DeleteQueue(ctx, &sqs.DeleteQueueInput{
		QueueUrl: urlResult.QueueUrl,
	})
	if err != nil {
		return err
	}

	log.Printf("[DeleteQueue] Deleted queue: %s", queueName)

	return nil
}
