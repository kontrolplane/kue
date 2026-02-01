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
	Queue kue.Queue
	Err   error
}

// MessagesLoadedMsg is sent when queue messages have been loaded.
type MessagesLoadedMsg struct {
	Messages []kue.Message
	Err      error
}

// QueueCreatedMsg is sent when a queue has been created.
type QueueCreatedMsg struct {
	QueueUrl string
	Err      error
}

// QueueDeletedMsg is sent when a queue has been deleted.
type QueueDeletedMsg struct {
	Err error
}

// RefreshTickMsg is sent periodically to trigger data refresh.
type RefreshTickMsg struct {
	Page string // Identifies which page requested the refresh
}

// LoadingMsg indicates a loading state has started.
type LoadingMsg struct {
	Loading bool
	Message string
}
