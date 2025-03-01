package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// ListQueues lists the SQS queues
func ListQueues(ctx context.Context) (queueUrls []string, err error) {

	// Create an SQS client
	client := createSqsClient(ctx)

	paginator := sqs.NewListQueuesPaginator(client, &sqs.ListQueuesInput{})
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			log.Printf("Couldn't get queues. Here's why: %v\n", err)
			break
		} else {
			queueUrls = append(queueUrls, output.QueueUrls...)
		}
	}

	if len(queueUrls) == 0 {
		return nil, fmt.Errorf("No queues found")
	}

	return queueUrls, nil
}

// ListQueuesByPrefix lists the SQS queues by a given prefix
func ListQueuesByPrefix() (queueUrls []string, err error) {
	return nil, nil
}
