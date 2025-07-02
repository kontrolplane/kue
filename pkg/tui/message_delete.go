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
    message  kue.Message
    selected int // 0 = abort, 1 = confirm
}

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {

    log.Println("[QueueMessageDeleteSwitchPage]")

    return m.SwitchPage(queueMessageDelete), nil
}

func (m model) QueueMessageDeleteView() string {

    dialog := lipgloss.NewStyle().
        Padding(1, 3)

    secondary := lipgloss.NewStyle().
        Foreground(lipgloss.Color("#ffffff")).
        Background(lipgloss.Color("#628049")).
        Padding(0, 3)

    primary := lipgloss.NewStyle().
        Foreground(lipgloss.Color("255")).
        Background(lipgloss.Color("240")).
        Padding(0, 3)

    messageID := lipgloss.NewStyle().Bold(true).Render(m.state.queueMessageDelete.message.MessageID)

    confirm := "yes"
    abort := "no"

    if m.state.queueMessageDelete.selected == 0 {
        abort = primary.Render(abort)
        confirm = secondary.Render(confirm)
    } else {
        confirm = primary.Render(confirm)
        abort = secondary.Render(abort)
    }

    buttons := lipgloss.JoinHorizontal(
        lipgloss.Center,
        abort,
        "    ",
        confirm,
    )

    d := lipgloss.JoinVertical(
        lipgloss.Center,
        "warning: message deletion",
        "",
        "are you sure you want to delete message: "+messageID+" ?",
        "",
        buttons,
    )

    return dialog.Render(d)
}

func (m model) switchMessageDeleteOption() (model, tea.Cmd) {
    m.state.queueMessageDelete.selected = (m.state.queueMessageDelete.selected + 1) % 2
    return m, nil
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Left):
            m, cmd = m.switchMessageDeleteOption()
        case key.Matches(msg, m.keys.Right):
            m, cmd = m.switchMessageDeleteOption()
        case key.Matches(msg, m.keys.View):
            if m.state.queueMessageDelete.selected == 0 {
                // abort
                m.state.queueMessageDelete.selected = 0
                return m.QueueDetailsSwitchPage(msg)
            }

            // confirm deletion
            err := kue.DeleteMessage(
                m.client,
                m.context,
                m.state.queueDetails.queue.Url,
                m.state.queueMessageDelete.message.ReceiptHandle,
            )
            if err != nil {
                m.error = fmt.Sprintf("Error deleting message: %v", err)
            }
            m.state.queueMessageDelete.selected = 0
            return m.QueueDetailsSwitchPage(msg)
        case key.Matches(msg, m.keys.Quit):
            m.state.queueMessageDelete.selected = 0
            return m.QueueDetailsSwitchPage(msg)
        }
    }

    return m, cmd
}
