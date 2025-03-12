package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// fetchContext loads the default AWS configuration using the AWS SDK for Go.
// It returns the loaded aws.Config and an error if the configuration could not be loaded.
//
// Parameters:
//
//	ctx - The context from which to fetch the configuration.
//
// Returns:
//
//   - aws.Config: The loaded AWS configuration.
//   - error: An error if the configuration could not be loaded.
func fetchContext(ctx context.Context) (aws.Config, error) {
	config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, err
	}

	return config, nil
}

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
func CreateSqsClient(ctx context.Context) (*sqs.Client, error) {
	config, err := fetchContext(ctx)
	if err != nil {
		return nil, err
	}

	return sqs.NewFromConfig(config), nil
}
