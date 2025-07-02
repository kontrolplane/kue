package kue

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueueInput contains the parameters that can be supplied when creating
// a queue from the TUI. All fields are optional except Name.
// A zero value indicates the default AWS SQS value should be used.
// Only a subset of attributes are exposed for now – this can be extended later
// without changing the behaviour of the call site.
type CreateQueueInput struct {
	Name                   string
	VisibilityTimeout      int32 // seconds (0 – 43200)
	MessageRetentionPeriod int32 // seconds (60 – 1209600)
	DelaySeconds           int32 // seconds (0 – 900)
	ReceiveWaitTimeSeconds int32 // seconds (0 – 20)
}

// CreateQueue creates a new SQS queue with the supplied attributes and returns
// the queue URL. It wraps sqs.CreateQueue and translates non-zero values into
// the Attribute map expected by the AWS SDK.
func CreateQueue(client *sqs.Client, ctx context.Context, in CreateQueueInput) (string, error) {
	if in.Name == "" {
		return "", fmt.Errorf("queue name must be provided")
	}

	attrs := make(map[string]string)
	// Only set attributes that have non-zero values so that we rely on the AWS
	// defaults where the user has not provided an override.
	if in.VisibilityTimeout != 0 {
		attrs["VisibilityTimeout"] = fmt.Sprintf("%d", in.VisibilityTimeout)
	}
	if in.MessageRetentionPeriod != 0 {
		attrs["MessageRetentionPeriod"] = fmt.Sprintf("%d", in.MessageRetentionPeriod)
	}
	if in.DelaySeconds != 0 {
		attrs["DelaySeconds"] = fmt.Sprintf("%d", in.DelaySeconds)
	}
	if in.ReceiveWaitTimeSeconds != 0 {
		attrs["ReceiveMessageWaitTimeSeconds"] = fmt.Sprintf("%d", in.ReceiveWaitTimeSeconds)
	}

	out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName:  &in.Name,
		Attributes: attrs,
	})
	if err != nil {
		return "", err
	}
	return *out.QueueUrl, nil
}

