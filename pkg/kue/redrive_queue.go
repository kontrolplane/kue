package kue

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// MessageMoveTaskStatus represents the status of a message move task.
type MessageMoveTaskStatus struct {
	TaskHandle                        string
	Status                            string
	SourceArn                         string
	DestinationArn                    string
	ApproximateNumberOfMessagesMoved  int64
	ApproximateNumberOfMessagesToMove int64
	FailureReason                     string
}

// StartMessageMoveTask starts a redrive task moving messages from the
// source (DLQ) ARN back to the destination queue.
func StartMessageMoveTask(client *sqs.Client, ctx context.Context, sourceArn string, destinationArn string) (string, error) {
	input := &sqs.StartMessageMoveTaskInput{
		SourceArn:      &sourceArn,
		DestinationArn: &destinationArn,
	}

	result, err := client.StartMessageMoveTask(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to start message move task: %w", err)
	}

	taskHandle := ""
	if result.TaskHandle != nil {
		taskHandle = *result.TaskHandle
	}

	log.Printf("[StartMessageMoveTask] Started redrive from %s, task handle: %s", sourceArn, taskHandle)
	return taskHandle, nil
}

// ListMessageMoveTasks returns the active/recent message move tasks for the given source ARN.
func ListMessageMoveTasks(client *sqs.Client, ctx context.Context, sourceArn string) ([]MessageMoveTaskStatus, error) {
	result, err := client.ListMessageMoveTasks(ctx, &sqs.ListMessageMoveTasksInput{
		SourceArn: &sourceArn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list message move tasks: %w", err)
	}

	var tasks []MessageMoveTaskStatus
	for _, t := range result.Results {
		task := MessageMoveTaskStatus{
			ApproximateNumberOfMessagesMoved: t.ApproximateNumberOfMessagesMoved,
		}
		if t.ApproximateNumberOfMessagesToMove != nil {
			task.ApproximateNumberOfMessagesToMove = *t.ApproximateNumberOfMessagesToMove
		}
		if t.TaskHandle != nil {
			task.TaskHandle = *t.TaskHandle
		}
		if t.SourceArn != nil {
			task.SourceArn = *t.SourceArn
		}
		if t.DestinationArn != nil {
			task.DestinationArn = *t.DestinationArn
		}
		if t.Status != nil {
			task.Status = *t.Status
		}
		if t.FailureReason != nil {
			task.FailureReason = *t.FailureReason
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
