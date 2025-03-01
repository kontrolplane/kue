package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Queue struct {
	Protocol          string
	ServiceEndpoint   string
	AccountIdentifier string
	Name              string
}

// ListQueues lists the SQS queues
func ListQueues(ctx context.Context) (queues []Queue, err error) {

	// Create an SQS client
	client := createSqsClient(ctx)

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
		queue, err := extractQueueStruct(queueUrl)
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

// ListQueuesByPrefix lists the SQS queues by a given prefix
func ListQueuesByPrefix() (queueUrls []string, err error) {
	return nil, nil
}

// extractQueueStruct extracts the queue struct from a queue URL
func extractQueueStruct(queueUrl string) (Queue, error) {
	parts := strings.Split(queueUrl, "/")
	if len(parts) < 5 {
		return Queue{}, fmt.Errorf("Invalid queue URL: %s", queueUrl)
	}

	return Queue{
		Protocol:          strings.TrimSuffix(parts[0], ":"),
		ServiceEndpoint:   parts[2],
		AccountIdentifier: parts[3],
		Name:              parts[4],
	}, nil
}
