package sqs

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// ListQueues lists the SQS queues
func ListQueues(client *sqs.Client, ctx context.Context) (queues []Queue, err error) {

	var queueUrls []string

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

	for _, queueUrl := range queueUrls {
		queue, err := formatQueueName(queueUrl)
		if err != nil {
			log.Printf("Couldn't extract queue struct. Here's why: %v\n", err)
			continue
		}
		queues = append(queues, queue)
	}

	if len(queues) == 0 {
		return nil, fmt.Errorf("No queues found")
	}

	return queues, nil
}
