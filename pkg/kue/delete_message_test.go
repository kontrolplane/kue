package kue

import (
    "context"
    "errors"
    "testing"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

type mockSQSClient struct{
    deleteErr error
}

func (m mockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
    return &sqs.DeleteMessageOutput{}, m.deleteErr
}

func TestDeleteMessageSuccess(t *testing.T) {
    client := mockSQSClient{}
    if err := DeleteMessage(client, context.Background(), "url", "rh"); err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
}

func TestDeleteMessageError(t *testing.T) {
    client := mockSQSClient{deleteErr: errors.New("boom")}
    if err := DeleteMessage(client, context.Background(), "url", "rh"); err == nil {
        t.Fatalf("expected error, got nil")
    }
}
