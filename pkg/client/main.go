package client

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// AWSInfo holds AWS configuration information for display.
type AWSInfo struct {
	Profile string
	Region  string
}

// fetchContext loads the default AWS configuration using the AWS SDK for Go.
// It returns the loaded aws.Config and an error if the configuration could not be loaded.
func fetchContext(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, err
	}

	return cfg, nil
}

// CreateSqsClient creates and returns a new Amazon SQS client using the provided context.
// It also returns AWS configuration info (profile, region) for display purposes.
func CreateSqsClient(ctx context.Context) (*sqs.Client, AWSInfo, error) {
	cfg, err := fetchContext(ctx)
	if err != nil {
		return nil, AWSInfo{}, err
	}

	// Get profile from environment (AWS SDK doesn't expose it directly)
	profile := os.Getenv("AWS_PROFILE")
	if profile == "" {
		profile = "default"
	}

	info := AWSInfo{
		Profile: profile,
		Region:  cfg.Region,
	}

	return sqs.NewFromConfig(cfg), info, nil
}
