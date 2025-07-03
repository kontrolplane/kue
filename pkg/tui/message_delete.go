package tui

import (
    "fmt"
    "log"

    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/lipgloss"

    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueMessageDeleteState struct {
    queue    kue.Queue
    message  kue.Message
    selected int // 0 = move to DLQ, 1 = delete permanently
}

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    log.Println("[QueueMessageDeleteSwitchPage]")
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

    // Render queue & message information
    queueName := lipgloss.NewStyle().Bold(true).Render(m.state.queueMessageDelete.queue.Name)
    messageId := lipgloss.NewStyle().Italic(true).Render(m.state.queueMessageDelete.message.MessageID)

    moveToDLQ := "move to dlq"
    deletePermanent := "delete"

    if m.state.queueMessageDelete.selected == 0 {
        moveToDLQ = primary.Render(moveToDLQ)
        deletePermanent = secondary.Render(deletePermanent)
    } else {
        moveToDLQ = secondary.Render(moveToDLQ)
        deletePermanent = primary.Render(deletePermanent)
    }

    buttons := lipgloss.JoinHorizontal(lipgloss.Center, moveToDLQ, "    ", deletePermanent)

    d := lipgloss.JoinVertical(
        lipgloss.Center,
        "message action",
        "",
        fmt.Sprintf("queue: %s", queueName),
        fmt.Sprintf("message id: %s", messageId),
        "",
        buttons,
    )

    return dialog.Render(d)
}

func (m model) switchMessageAction() (model, tea.Cmd) {
    m.state.queueMessageDelete.selected = (m.state.queueMessageDelete.selected + 1) % 2
    return m, nil
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Left):
            m, cmd = m.switchMessageAction()
        case key.Matches(msg, m.keys.Right):
            m, cmd = m.switchMessageAction()
        case key.Matches(msg, m.keys.View):
            // Enter pressed on chosen option
            if m.state.queueMessageDelete.selected == 0 {
                // move to DLQ
                if err := kue.MoveMessageToDLQ(m.context, m.client, m.state.queueMessageDelete.queue.Url, m.state.queueMessageDelete.message); err != nil {
                    m.error = fmt.Sprintf("Error moving message to DLQ: %v", err)
                }
            } else {
                // delete permanently
                if err := kue.DeleteMessage(m.context, m.client, m.state.queueMessageDelete.queue.Url, m.state.queueMessageDelete.message.ReceiptHandle); err != nil {
                    m.error = fmt.Sprintf("Error deleting message: %v", err)
                }
            }
            // Reset selection & refresh queue details
            m.state.queueMessageDelete.selected = 0
            return m.QueueDetailsSwitchPage(msg)
        case key.Matches(msg, m.keys.Quit):
            // Abort
            m.state.queueMessageDelete.selected = 0
            return m.QueueDetailsSwitchPage(msg)
        }
    }

    return m, cmd
}
