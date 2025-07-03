package tui

import kuepkg "github.com/kontrolplane/kue/pkg/kue"

// QueueCreateResultMsg is emitted when the CreateQueue operation finishes.
// If Err is non-nil, queue creation failed.
// On success Queue will contain the created queue.
// This message is handled in root.Update().
//
// It enables optimistic UI where spinner stops and toast shows outcome.
//
// Defined in its own file to avoid import cycles: UI layer only depends on kue package.
type QueueCreateResultMsg struct {
    Queue kuepkg.Queue
    Err   error
}

type ToastClearMsg struct{}
