// Package commands provides tea.Cmd factories for async operations.
package commands

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/messages"
)

// RefreshInterval is the default interval for auto-refresh.
const RefreshInterval = 30 * time.Second

// LoadQueues creates a command to load all queues with their attributes.
func LoadQueues(ctx context.Context, client *sqs.Client) tea.Cmd {
	return func() tea.Msg {
		queues, err := kue.ListQueuesUrls(client, ctx)
		if err != nil {
			return messages.QueuesLoadedMsg{Queues: nil, Err: err}
		}

		// Fetch attributes for each queue
		for i, queue := range queues {
			q, err := kue.FetchQueueAttributes(client, ctx, queue.Url)
			if err != nil {
				return messages.QueuesLoadedMsg{Queues: nil, Err: err}
			}
			queues[i] = q
		}

		return messages.QueuesLoadedMsg{Queues: queues, Err: nil}
	}
}

// LoadQueueAttributes creates a command to load attributes for a specific queue.
func LoadQueueAttributes(ctx context.Context, client *sqs.Client, queueUrl string) tea.Cmd {
	return func() tea.Msg {
		queue, err := kue.FetchQueueAttributes(client, ctx, queueUrl)
		return messages.QueueAttributesLoadedMsg{Queue: queue, Err: err}
	}
}

// LoadMessages creates a command to load messages from a queue.
func LoadMessages(ctx context.Context, client *sqs.Client, queueUrl string, maxMessages int32) tea.Cmd {
	return func() tea.Msg {
		msgs, err := kue.FetchQueueMessages(client, ctx, queueUrl, maxMessages)
		return messages.MessagesLoadedMsg{Messages: msgs, Err: err}
	}
}

// CreateQueue creates a command to create a new queue.
func CreateQueue(ctx context.Context, client *sqs.Client, config kue.QueueConfig) tea.Cmd {
	return func() tea.Msg {
		url, err := kue.CreateQueue(client, ctx, config)
		queueUrl := ""
		if url != nil {
			queueUrl = *url
		}
		return messages.QueueCreatedMsg{QueueUrl: queueUrl, Err: err}
	}
}

// DeleteQueue creates a command to delete a queue.
func DeleteQueue(ctx context.Context, client *sqs.Client, queueName string) tea.Cmd {
	return func() tea.Msg {
		err := kue.DeleteQueue(client, ctx, queueName)
		return messages.QueueDeletedMsg{Err: err}
	}
}

// DeleteQueues creates a command to delete multiple queues.
func DeleteQueues(ctx context.Context, client *sqs.Client, queues []kue.Queue) tea.Cmd {
	return func() tea.Msg {
		for _, q := range queues {
			if err := kue.DeleteQueue(client, ctx, q.Name); err != nil {
				return messages.QueueDeletedMsg{Err: err}
			}
		}
		return messages.QueueDeletedMsg{Err: nil}
	}
}

// DeleteMessage creates a command to delete a message from a queue.
func DeleteMessage(ctx context.Context, client *sqs.Client, queueUrl string, receiptHandle string) tea.Cmd {
	return func() tea.Msg {
		err := kue.DeleteMessage(client, ctx, queueUrl, receiptHandle)
		return messages.MessageDeletedMsg{Err: err}
	}
}

// DeleteMessages creates a command to delete multiple messages from a queue.
func DeleteMessages(ctx context.Context, client *sqs.Client, queueUrl string, msgs []kue.Message) tea.Cmd {
	return func() tea.Msg {
		for _, msg := range msgs {
			if err := kue.DeleteMessage(client, ctx, queueUrl, msg.ReceiptHandle); err != nil {
				return messages.MessageDeletedMsg{Err: err}
			}
		}
		return messages.MessageDeletedMsg{Err: nil}
	}
}

// SendMessage creates a command to send a message to a queue.
func SendMessage(ctx context.Context, client *sqs.Client, input kue.SendMessageInput) tea.Cmd {
	return func() tea.Msg {
		err := kue.SendMessage(client, ctx, input)
		return messages.MessageCreatedMsg{Err: err}
	}
}

// RefreshTick creates a command that sends a refresh message after the given duration.
func RefreshTick(d time.Duration, page string) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return messages.RefreshTickMsg{Page: page}
	})
}

// ScheduleRefresh creates a command to schedule the next refresh for a page.
func ScheduleRefresh(page string) tea.Cmd {
	return RefreshTick(RefreshInterval, page)
}
