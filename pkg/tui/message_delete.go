package tui

import (
    "fmt"
    "log"
    "strings"

    "github.com/charmbracelet/lipgloss"

    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

type messageDeleteState struct {
    message kue.Message
}

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
    log.Println("[QueueMessageDeleteSwitchPage]")
    return m.SwitchPage(queueMessageDelete), nil
}

func (m model) QueueMessageDeleteView() string {
    dialog := lipgloss.NewStyle().Padding(1, 3)

    // Build excerpt of message body (first 80 characters)
    body := m.state.queueMessageDelete.message.Body
    excerptLength := 80
    if len(body) > excerptLength {
        body = body[:excerptLength] + "…"
    }
    excerpt := fmt.Sprintf("\"%s\"", strings.ReplaceAll(body, "\n", " "))

    prompt := "Delete this message? Type 'y' to confirm, any other key to cancel."

    d := lipgloss.JoinVertical(
        lipgloss.Left,
        "warning: message deletion",
        "",
        excerpt,
        "",
        prompt,
    )

    return dialog.Render(d)
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        log.Printf("[QueueMessageDeleteUpdate] Key pressed: %s", msg.String())
        switch {
        case msg.String() == "y" || msg.String() == "Y":
            // Perform deletion
            err := kue.DeleteMessage(m.client, m.context, m.state.queueDetails.queue.Url, m.state.queueMessageDelete.message.ReceiptHandle)
            if err != nil {
                m.error = fmt.Sprintf("Error deleting message: %v", err)
            }
            // Refresh messages list after deletion
            messages, err := kue.FetchQueueMessages(m.client, m.context, m.state.queueDetails.queue.Url, 10)
            if err != nil {
                m.error = fmt.Sprintf("Error fetching queue message(s): %v", err)
            }
            m.state.queueDetails.messages = messages
            m.state.queueDetails.selected = 0
            return m.QueueDetailsSwitchPage(msg)
        default:
            // cancel
            return m.QueueDetailsSwitchPage(msg)
        }
    }
    return m, nil
}
