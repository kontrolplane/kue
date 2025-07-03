package tui

import (
    "fmt"

    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueMessageDeleteState struct {
    message  kue.Message
    selected int // 0 = abort, 1 = confirm
}

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
    return m.SwitchPage(queueMessageDelete), nil
}

func (m model) QueueMessageDeleteView() string {
    dialog := lipgloss.NewStyle().Padding(1, 3)

    secondary := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#ffffff")).
        Background(lipgloss.Color("#628049")).
        Padding(0, 3)

    primary := lipgloss.NewStyle().
        Foreground(lipgloss.Color("255")).
        Background(lipgloss.Color("240")).
        Padding(0, 3)

    confirm := "yes"
    abort := "no"

    if m.state.queueMessageDelete.selected == 1 {
        confirm = primary.Render(confirm)
        abort = secondary.Render(abort)
    } else {
        confirm = secondary.Render(confirm)
        abort = primary.Render(abort)
    }

    buttons := lipgloss.JoinHorizontal(
        lipgloss.Center,
        abort,
        "    ",
        confirm,
    )

    body := lipgloss.NewStyle().Width(50)

    d := lipgloss.JoinVertical(
        lipgloss.Center,
        "warning: message deletion",
        "",
        fmt.Sprintf("delete message: %s…?", truncate(m.state.queueMessageDelete.message.MessageID, 25)),
        "",
        buttons,
    )

    return dialog.Render(body.Render(d))
}

func (m model) switchMessageDeleteOption() (model, tea.Cmd) {
    m.state.queueMessageDelete.selected = (m.state.queueMessageDelete.selected + 1) % 2
    return m, nil
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Left):
            m, cmd = m.switchMessageDeleteOption()
        case key.Matches(msg, m.keys.Right):
            m, cmd = m.switchMessageDeleteOption()
        case key.Matches(msg, m.keys.View): // enter to choose
            if m.state.queueMessageDelete.selected == 0 {
                // abort
                m.state.queueMessageDelete.selected = 0
                return m.QueueDetailsSwitchPage(msg)
            }

            // confirm delete
            err := kue.DeleteMessage(
                m.client,
                m.context,
                m.state.queueDetails.queue.Url,
                m.state.queueMessageDelete.message.ReceiptHandle,
            )
            if err != nil {
                m.error = fmt.Sprintf("Error deleting message: %v", err)
            }

            // refresh messages list after deletion
            refreshed, err := kue.FetchQueueMessages(m.client, m.context, m.state.queueDetails.queue.Url, 10)
            if err != nil {
                m.error = fmt.Sprintf("Error fetching queue message(s): %v", err)
            }
            m.state.queueDetails.messages = refreshed
            m.state.queueMessageDelete.selected = 0
            return m.QueueDetailsSwitchPage(msg)
        case key.Matches(msg, m.keys.Quit):
            m.state.queueMessageDelete.selected = 0
            return m.QueueDetailsSwitchPage(msg)
        }
    }
    return m, cmd
}

// helper to truncate a string with ellipsis if it exceeds length
func truncate(s string, max int) string {
    if len(s) <= max {
        return s
    }
    if max <= 3 {
        return s[:max]
    }
    return s[:max-3] + "…"
}
