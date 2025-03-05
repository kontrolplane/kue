package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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
