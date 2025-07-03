package kue

import (
    "context"
    "fmt"
    "strings"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
)

// DefaultDelimiter is the marker that separates multiple message bodies when the
// user wants to send a batch in the message creation view. We purposefully use a
// newline terminated triple dash ("---") that is unlikely to appear in JSON
// payloads and can still be typed easily. The expected layout is therefore:
//
//     {"foo":"bar"}
//     ---
//     {"hello":"world"}
//
// Empty fragments produced by leading/trailing delimiters are ignored.
const DefaultDelimiter = "---"

// SplitBodies takes the raw user input and returns a slice containing the
// individual message bodies. Leading and trailing whitespace on each body is
// trimmed. Empty bodies are discarded.
func SplitBodies(raw string, delimiter string) []string {
    if delimiter == "" {
        delimiter = DefaultDelimiter
    }
    // We normalise Windows line-endings just in case
    normalised := strings.ReplaceAll(raw, "\r\n", "\n")
    parts := strings.Split(normalised, delimiter)
    bodies := make([]string, 0, len(parts))
    for _, p := range parts {
        body := strings.TrimSpace(p)
        if body == "" {
            continue
        }
        bodies = append(bodies, body)
    }
    return bodies
}

// SendMessage sends a single message body to the given queue URL with the
// specified delivery delay. A delaySeconds value of 0 means no delay.
func SendMessage(client *sqs.Client, ctx context.Context, queueURL string, body string, delaySeconds int32) error {
    input := &sqs.SendMessageInput{
        QueueUrl:    aws.String(queueURL),
        MessageBody: aws.String(body),
    }

    if delaySeconds > 0 {
        input.DelaySeconds = delaySeconds
    }

    _, err := client.SendMessage(ctx, input)
    if err != nil {
        return fmt.Errorf("failed sending message: %w", err)
    }
    return nil
}

// SendMessages iterates over the provided message bodies and sends them one by
// one to the queue. It returns a channel that emits progress updates (number of
// successfully sent messages) until the operation completes. The channel is
// closed once every message has been processed.
func SendMessages(client *sqs.Client, ctx context.Context, queueURL string, bodies []string, delaySeconds int32) <-chan int {
    progress := make(chan int)
    go func() {
        defer close(progress)
        var sent int
        for _, body := range bodies {
            if ctx.Err() != nil {
                // Context cancelled; abort early
                return
            }
            if err := SendMessage(client, ctx, queueURL, body, delaySeconds); err == nil {
                sent++
                progress <- sent
            } else {
                // We still publish progress but as negative number to indicate failure
                progress <- -1
            }
        }
    }()
    return progress
}
