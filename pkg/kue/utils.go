package kue

import "github.com/aws/aws-sdk-go-v2/aws"

// awsString returns a pointer to the provided string.
func awsString(s string) *string { return aws.String(s) }
