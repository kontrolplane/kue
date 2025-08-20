package tui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueMessageDetailsState struct {
    message kue.Message
}

func (m model) QueueMessageDetailsSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m.SwitchPage(queueMessageDetails), nil
}

func renderMessageDetails(message kue.Message) string {
    var b strings.Builder

    // Basic fields
    write := func(key, value string) {
        if value != "" {
            fmt.Fprintf(&b, "%s: %s\n", key, value)
        }
    }

    write("message id", message.MessageID)
    write("receipt handle", message.ReceiptHandle)
    write("md5 of body", message.MD5OfBody)
    write("sent timestamp", message.SentTimestamp)
    write("first receive time", message.FirstReceiveTime)
    write("receive count", message.ReceiveCount)
    write("message group id", message.MessageGroupID)
    write("message deduplication id", message.MessageDeduplicationID)
    write("sequence number", message.SequenceNumber)

    // Body (might be large, keep after attributes)
    if message.Body != "" {
        fmt.Fprintf(&b, "body:\n%s\n", message.Body)
    }

    // Attributes
    if len(message.Attributes) > 0 {
        b.WriteString("attributes:\n")
        for k, v := range message.Attributes {
            fmt.Fprintf(&b, "  %s: %s\n", k, v)
        }
    }

    // MessageAttributes
    if len(message.MessageAttributes) > 0 {
        b.WriteString("message attributes:\n")
        for k, v := range message.MessageAttributes {
            fmt.Fprintf(&b, "  %s: %s\n", k, v)
        }
    }

    return b.String()
}

func (m model) QueueMessageDetailsView() string {
    // Use lipgloss to ensure wrapping within window width
    content := renderMessageDetails(m.state.queueMessageDetails.message)

    width := m.width - 4 // consider padding
    if width < 0 {
        width = 0
    }
    return lipgloss.NewStyle().Width(width).Render(content)
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit):
            // Return to queue details page
            return m.QueueDetailsSwitchPage(msg)
        }
    }
    return m, cmd
}
