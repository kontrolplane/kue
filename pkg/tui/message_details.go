package tui

import (
    "fmt"

    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueMessageDetailsState keeps the currently selected message to show
// in the message details page.
type queueMessageDetailsState struct {
    message kue.Message
}

func (m model) QueueMessageDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    return m.SwitchPage(queueMessageDetails), nil
}

func (m model) QueueMessageDetailsView() string {
    msg := m.state.queueMessageDetails.message

    if msg.MessageID == "" {
        return "no message selected"
    }

    // Produce a simple multi-line view with the most important information
    return fmt.Sprintf(`Message details

ID:   %s
Sent: %s
Size: %d bytes

Body:
%s`,
        msg.MessageID,
        msg.SentTimestamp,
        len(msg.Body),
        msg.Body,
    )
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
    switch t := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(t, m.keys.Quit):
            // Go back to the queue details page
            return m.QueueDetailsSwitchPage(t)
        case key.Matches(t, m.keys.View):
            // Treat an additional enter as going back
            return m.QueueDetailsSwitchPage(t)
        }
    }
    return m, nil
}
