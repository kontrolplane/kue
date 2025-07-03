package commands

import (
    "context"
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/aws/aws-sdk-go-v2/service/sqs"

    kue "github.com/kontrolplane/kue/pkg/kue"
    tui "github.com/kontrolplane/kue/pkg/tui"
)

func CreateQueueCmd(client *sqs.Client, ctx context.Context, name string, attrs map[string]string, tags map[string]string) tea.Cmd {
    return func() tea.Msg {
        q, err := kue.CreateQueue(client, ctx, name, attrs, tags)
        return tui.QueueCreateResultMsg{Queue: q, Err: err}
    }
}

func ClearToastCmd(seconds time.Duration) tea.Cmd {
    return tea.Tick(seconds, func(t time.Time) tea.Msg { return tui.ToastClearMsg{} })
}
