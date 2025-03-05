package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// createSqsClient creates and returns a new Amazon SQS client using the provided context.
// It fetches the configuration from the context and uses it to initialize the SQS client.
// If there is an error fetching the configuration, it returns nil.
//
// Parameters:
//
//	ctx - The context from which to fetch the configuration.
//
// Returns:
//
//   - *sqs.Client - A new SQS client initialized with the fetched configuration, or nil if an error occurs.
func createSqsClient(ctx context.Context) *sqs.Client {
	config, err := fetchContext(ctx)
	if err != nil {
		return nil
	}

	return sqs.NewFromConfig(config)
}
