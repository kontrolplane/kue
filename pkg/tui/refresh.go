package tui

import (
    "context"
    "os"
    "strconv"
    "time"

    "github.com/aws/aws-sdk-go-v2/service/sqs"
    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

const defaultRefreshSeconds = 5

// readRefreshInterval returns refresh interval from env var KUE_REFRESH_INTERVAL (seconds) or the default.
func readRefreshInterval() time.Duration {
    if s, ok := os.LookupEnv("KUE_REFRESH_INTERVAL"); ok {
        if v, err := strconv.Atoi(s); err == nil && v > 0 {
            return time.Duration(v) * time.Second
        }
    }
    return time.Duration(defaultRefreshSeconds) * time.Second
}

// startQueuePolling starts background goroutine which periodically fetches queue attributes.
func startQueuePolling(ctx context.Context, client *sqs.Client, interval time.Duration) <-chan []kue.Queue {
    // We can’t import the AWS SQS type here to avoid cycles; expose *sqs.Client via interface.
    ch := make(chan []kue.Queue)
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        for {
            // first tick immediately
            var trigger <-chan time.Time
            if interval > 0 {
                trigger = ticker.C
            }
            select {
            case <-ctx.Done():
                close(ch)
                return
            case <-trigger:
                queues, _ := kue.ListQueuesUrls(client, ctx)
                for i, q := range queues {
                    q, _ = kue.FetchQueueAttributes(client, ctx, q.Url)
                    queues[i] = q
                }
                ch <- queues
            }
        }
    }()
    return ch
}

// readQueuesCmd returns a Bubble Tea command that waits for a slice from channel.
func readQueuesCmd(ch <-chan []kue.Queue) tea.Cmd {
    return func() tea.Msg {
        queues, ok := <-ch
        if !ok {
            return nil
        }
        return UpdateQueuesMsg{Queues: queues}
    }
}
