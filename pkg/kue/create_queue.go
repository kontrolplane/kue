package kue

import (
    "context"
    "encoding/json"
    "fmt"
    "strconv"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// CreateQueueInput represents optional configuration for a queue to be created.
// Fields may be left at their zero value to omit them from the CreateQueue call.
// All durations are expressed in seconds, matching the AWS SQS API.
//
// Only attributes with non-zero/empty values are sent to AWS to keep behaviour
// consistent with the console and to avoid overriding AWS defaults.
//
// VisibilityTimeout: 0 – 43200 (12 hours)
// MessageRetentionPeriod: 60 – 1209600 (14 days)
// DeliveryDelay: 0 – 900 (15 minutes) (not part of advanced section but kept for completeness)
// MaximumMessageSize: 1024 – 262144 (256 KiB)
// ReceiveMessageWaitTimeSeconds: 0 – 20
//
// Dead-letter queues require a RedrivePolicy which references the DLQ ARN and a
// maximum receive count. We expose DlqArn and DlqMaxReceiveCount (defaults to 5
// if DlqArn is provided but count is 0).
//
// Encryption (KMS) is enabled when KmsKeyID is non-empty. Supplying the empty
// string leaves encryption disabled (SQS uses SSE‐SQS managed key).
//
// Note: This struct intentionally lives in the kue package (not the TUI)
// so it can be reused by non-interactive code and unit tests.
//
//go:generate go test -v ./...
//
type CreateQueueInput struct {
    Name                         string
    VisibilityTimeout            int32  // seconds – optional (0 to omit)
    MessageRetentionPeriod       int32  // seconds – optional (0 to omit)
    DlqArn                       string // optional – when provided we create RedrivePolicy
    DlqMaxReceiveCount           int32  // default 5 when DlqArn != ""
    KmsKeyID                     string // optional customer managed KMS key
    DeliveryDelay                int32  // seconds – optional
    MaximumMessageSize           int32  // bytes – optional
    ReceiveMessageWaitTime       int32  // seconds – optional
}

// CreateQueue creates an SQS queue applying only the non-zero/empty attributes
// supplied in opts. On success it returns the queue URL.
func CreateQueue(client *sqs.Client, ctx context.Context, opts CreateQueueInput) (string, error) {
    if opts.Name == "" {
        return "", fmt.Errorf("queue name cannot be empty")
    }

    attrs := make(map[string]string)

    if opts.VisibilityTimeout > 0 {
        attrs["VisibilityTimeout"] = strconv.Itoa(int(opts.VisibilityTimeout))
    }
    if opts.MessageRetentionPeriod > 0 {
        attrs["MessageRetentionPeriod"] = strconv.Itoa(int(opts.MessageRetentionPeriod))
    }
    if opts.DeliveryDelay > 0 {
        attrs["DelaySeconds"] = strconv.Itoa(int(opts.DeliveryDelay))
    }
    if opts.MaximumMessageSize > 0 {
        attrs["MaximumMessageSize"] = strconv.Itoa(int(opts.MaximumMessageSize))
    }
    if opts.ReceiveMessageWaitTime > 0 {
        attrs["ReceiveMessageWaitTimeSeconds"] = strconv.Itoa(int(opts.ReceiveMessageWaitTime))
    }
    if opts.KmsKeyID != "" {
        attrs["KmsMasterKeyId"] = opts.KmsKeyID
    }

    // Dead-letter queue handling (RedrivePolicy)
    if opts.DlqArn != "" {
        maxReceive := opts.DlqMaxReceiveCount
        if maxReceive == 0 {
            maxReceive = 5
        }
        rp := map[string]interface{}{
            "deadLetterTargetArn": opts.DlqArn,
            "maxReceiveCount":     maxReceive,
        }
        if b, err := json.Marshal(rp); err == nil {
            attrs["RedrivePolicy"] = string(b)
        } else {
            return "", fmt.Errorf("marshalling redrive policy: %w", err)
        }
    }

    out, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
        QueueName:  &opts.Name,
        Attributes: attrs,
    })
    if err != nil {
        return "", fmt.Errorf("create queue: %w", err)
    }
    if out.QueueUrl == nil {
        return "", fmt.Errorf("create queue succeeded but returned nil URL")
    }
    return *out.QueueUrl, nil
}
