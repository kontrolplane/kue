// Package messages defines all the custom tea.Msg types for the TUI application.
package messages

import (
	"github.com/kontrolplane/kue/pkg/kue"
)

// QueuesLoadedMsg is sent when the queue list has been loaded.
type QueuesLoadedMsg struct {
	Queues []kue.Queue
	Err    error
}

// QueueAttributesLoadedMsg is sent when queue attributes have been fetched.
type QueueAttributesLoadedMsg struct {
	Url   string
	Queue kue.Queue
	Err   error
}

// MessagesLoadedMsg is sent when queue messages have been loaded.
type MessagesLoadedMsg struct {
	Url      string
	Messages []kue.Message
	Err      error
}

// QueueCreatedMsg is sent when a queue has been created.
type QueueCreatedMsg struct {
	Name     string
	QueueUrl string
	Err      error
}

// QueuesDeletedMsg lists the queues that were deleted; Err holds the failure that stopped the others.
type QueuesDeletedMsg struct {
	Names []string
	Err   error
}

// MessagesDeletedMsg lists the messages that were deleted; Err holds the failure that stopped the others.
type MessagesDeletedMsg struct {
	MessageIDs []string
	Err        error
}

// MessageCreatedMsg is sent when a message has been sent to a queue.
type MessageCreatedMsg struct {
	Queue string
	Err   error
}

// RefreshTickMsg triggers a background refresh. Ticks whose generation no longer
// matches the model's are dropped, so at most one refresh loop is active.
type RefreshTickMsg struct {
	Gen int
}

// QueueRedriveStartedMsg is sent when a DLQ redrive task has been started.
type QueueRedriveStartedMsg struct {
	TaskHandle string
	Err        error
}

// QueueRedriveStatusMsg is sent when the redrive task status has been polled.
type QueueRedriveStatusMsg struct {
	Tasks []kue.MessageMoveTaskStatus
	Err   error
}

// QueuePurgedMsg is sent when a queue has been purged.
type QueuePurgedMsg struct {
	Queue string
	Err   error
}

// ClipboardCopiedMsg is sent after a clipboard copy operation completes.
type ClipboardCopiedMsg struct {
	Text string
	Err  error
}

// StatusClearMsg is sent to clear the transient status message.
type StatusClearMsg struct {
	Gen int
}

// ClockTickMsg redraws the time since the last refresh.
type ClockTickMsg struct{}
