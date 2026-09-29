// Package commands provides tea.Cmd factories for async operations.
package commands

import (
	"context"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/messages"
)

// RefreshInterval is the interval for auto-refresh of the visible data. Loading messages receives
// them, which counts towards their receive count, so it stays well apart.
const RefreshInterval = 30 * time.Second

// MessageLimit is the number of messages received for the queue details, the most SQS returns.
const MessageLimit = 10

const requestTimeout = 30 * time.Second

// request runs fn in a tea.Cmd with the request timeout applied to ctx.
func request(ctx context.Context, fn func(context.Context) tea.Msg) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()
		return fn(ctx)
	}
}

// LoadQueues creates a command to load all queues with their attributes.
func LoadQueues(ctx context.Context, client *sqs.Client) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		queues, err := kue.ListQueuesUrls(client, ctx)
		if err != nil {
			return messages.QueuesLoadedMsg{Err: err}
		}
		for i, queue := range queues {
			q, err := kue.FetchQueueAttributes(client, ctx, queue.Url)
			if err != nil {
				return messages.QueuesLoadedMsg{Err: err}
			}
			queues[i] = q
		}
		return messages.QueuesLoadedMsg{Queues: queues}
	})
}

// LoadQueueAttributes creates a command to load attributes for a specific queue.
func LoadQueueAttributes(ctx context.Context, client *sqs.Client, queueUrl string) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		queue, err := kue.FetchQueueAttributes(client, ctx, queueUrl)
		return messages.QueueAttributesLoadedMsg{Url: queueUrl, Queue: queue, Err: err}
	})
}

// LoadMessages creates a command to load messages from a queue.
func LoadMessages(ctx context.Context, client *sqs.Client, queueUrl string) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		msgs, err := kue.FetchQueueMessages(client, ctx, queueUrl, MessageLimit)
		return messages.MessagesLoadedMsg{Url: queueUrl, Messages: msgs, Err: err}
	})
}

// LoadQueueDetails loads the attributes and messages of a queue.
func LoadQueueDetails(ctx context.Context, client *sqs.Client, queueUrl string) tea.Cmd {
	return tea.Batch(LoadQueueAttributes(ctx, client, queueUrl), LoadMessages(ctx, client, queueUrl))
}

// CreateQueue creates a command to create a new queue.
func CreateQueue(ctx context.Context, client *sqs.Client, config kue.QueueConfig) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		url, err := kue.CreateQueue(client, ctx, config)
		queueUrl := ""
		if url != nil {
			queueUrl = *url
		}
		return messages.QueueCreatedMsg{Name: config.Name, QueueUrl: queueUrl, Err: err}
	})
}

// DeleteQueues creates a command to delete queues, stopping at the first failure.
func DeleteQueues(ctx context.Context, client *sqs.Client, names []string) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		var deleted []string
		for _, name := range names {
			if err := kue.DeleteQueue(client, ctx, name); err != nil {
				return messages.QueuesDeletedMsg{Names: deleted, Err: err}
			}
			deleted = append(deleted, name)
		}
		return messages.QueuesDeletedMsg{Names: deleted}
	})
}

// DeleteMessages creates a command to delete messages from a queue, stopping at the first failure.
func DeleteMessages(ctx context.Context, client *sqs.Client, queueUrl string, msgs []kue.Message) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		var deleted []string
		for _, msg := range msgs {
			if err := kue.DeleteMessage(client, ctx, queueUrl, msg.ReceiptHandle); err != nil {
				return messages.MessagesDeletedMsg{MessageIDs: deleted, Err: err}
			}
			deleted = append(deleted, msg.MessageID)
		}
		return messages.MessagesDeletedMsg{MessageIDs: deleted}
	})
}

// SendMessage creates a command to send a message to a queue.
func SendMessage(ctx context.Context, client *sqs.Client, queue string, input kue.SendMessageInput) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		err := kue.SendMessage(client, ctx, input)
		return messages.MessageCreatedMsg{Queue: queue, Err: err}
	})
}

// ScheduleRefresh sends a refresh tick of generation gen after the refresh interval.
func ScheduleRefresh(gen int) tea.Cmd {
	return tea.Tick(RefreshInterval, func(time.Time) tea.Msg {
		return messages.RefreshTickMsg{Gen: gen}
	})
}

// StartRedrive creates a command to start a DLQ redrive task.
func StartRedrive(ctx context.Context, client *sqs.Client, sourceArn string, destinationArn string) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		taskHandle, err := kue.StartMessageMoveTask(client, ctx, sourceArn, destinationArn)
		return messages.QueueRedriveStartedMsg{TaskHandle: taskHandle, Err: err}
	})
}

// ScheduleRedrivePoll creates a command that polls redrive status after a delay.
func ScheduleRedrivePoll(d time.Duration, ctx context.Context, client *sqs.Client, sourceArn string) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()
		tasks, err := kue.ListMessageMoveTasks(client, ctx, sourceArn)
		return messages.QueueRedriveStatusMsg{Tasks: tasks, Err: err}
	})
}

// PurgeQueue creates a command to purge all messages from a queue.
func PurgeQueue(ctx context.Context, client *sqs.Client, name, queueUrl string) tea.Cmd {
	return request(ctx, func(ctx context.Context) tea.Msg {
		err := kue.PurgeQueue(client, ctx, queueUrl)
		return messages.QueuePurgedMsg{Queue: name, Err: err}
	})
}

// CopyToClipboard copies text using the system clipboard. On failure the update
// loop falls back to OSC52, which also works over ssh.
func CopyToClipboard(text string) tea.Cmd {
	return func() tea.Msg {
		return messages.ClipboardCopiedMsg{Text: text, Err: clipboard.WriteAll(text)}
	}
}

// ScheduleClock ticks once after d, for a clock on screen that changes then.
func ScheduleClock(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return messages.ClockTickMsg{}
	})
}

// ClearStatusAfter sends a StatusClearMsg of generation gen after d.
func ClearStatusAfter(d time.Duration, gen int) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return messages.StatusClearMsg{Gen: gen}
	})
}
