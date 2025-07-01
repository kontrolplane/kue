package kue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// CreateQueueInput defines the user-controllable params for creating a queue.
type CreateQueueInput struct {
	Name                 string // required
	Fifo                 bool   // FIFO or Standard
	VisibilityTimeout    int    // in seconds
	MessageRetention     int    // in seconds
	DelaySeconds         int    // in seconds
}

// CreateQueue creates a new SQS queue and returns the new queue URL.
func CreateQueue(client *sqs.Client, ctx context.Context, input CreateQueueInput) (string, error) {
	attributes := make(map[string]string)
	if input.Fifo {
		attributes["FifoQueue"] = "true"
		if len(input.Name) < 5 || input.Name[len(input.Name)-5:] != ".fifo" {
			input.Name = input.Name + ".fifo"
		}
	}
	if input.VisibilityTimeout > 0 {
		attributes["VisibilityTimeout"] = fmt.Sprintf("%d", input.VisibilityTimeout)
	}
	if input.MessageRetention > 0 {
		attributes["MessageRetentionPeriod"] = fmt.Sprintf("%d", input.MessageRetention)
	}
	if input.DelaySeconds > 0 {
		attributes["DelaySeconds"] = fmt.Sprintf("%d", input.DelaySeconds)
	}
	o, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName:  &input.Name,
		Attributes: attributes,
	})
	if err != nil {
		return "", err
	}
	return *o.QueueUrl, nil
}
