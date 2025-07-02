package tui

import (
    "fmt"

    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/lipgloss"

    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueMessageDeleteState stores the context for the message deletion confirmation dialog.
// selected = 0 means "no", selected = 1 means "yes" (mirrors queueDelete implementation).
// We keep the queue URL handy because DeleteMessage needs it together with the receipt handle.
// The full Message is stored so we have access to its ReceiptHandle and can render details if we want.
type queueMessageDeleteState struct {
    queue    kue.Queue
    message  kue.Message
    selected int
}

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {

    // Nothing special to do besides switching the page; the queue and message
    // should already have been populated by the caller (QueueDetailsUpdate).
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

    // We only show the first 20 runes of the body to keep the dialog compact.
    previewBody := m.state.queueMessageDelete.message.Body
    if len(previewBody) > 20 {
        previewBody = previewBody[:20] + "…"
    }

    confirm := "yes"
    abort := "no"

    if m.state.queueMessageDelete.selected == 0 {
        // Currently on abort
        confirm = secondary.Render(confirm)
        abort = primary.Render(abort)
    } else {
        // Currently on confirm
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
        fmt.Sprintf("are you sure you want to delete message %s (body: %s) from queue %s?",
            m.state.queueMessageDelete.message.MessageID,
            previewBody,
            m.state.queueMessageDelete.queue.Name,
        ),
        "",
        buttons,
    )

    return dialog.Render(d)
}

// switchOption toggles the currently selected button (yes/no).
func (m model) switchOptionMessageDelete() (model, tea.Cmd) {
    m.state.queueMessageDelete.selected = (m.state.queueMessageDelete.selected + 1) % 2
    return m, nil
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Left):
            m, cmd = m.switchOptionMessageDelete()
        case key.Matches(msg, m.keys.Right):
            m, cmd = m.switchOptionMessageDelete()
        case key.Matches(msg, m.keys.View):
            if m.state.queueMessageDelete.selected == 0 {
                // abort
                m.state.queueMessageDelete.selected = 0
                return m.QueueDetailsSwitchPage(msg)
            } else {
                // confirm deletion
                if err := kue.DeleteMessage(
                    m.client,
                    m.context,
                    m.state.queueMessageDelete.queue.Url,
                    m.state.queueMessageDelete.message.ReceiptHandle,
                ); err != nil {
                    m.error = fmt.Sprintf("Error deleting message: %v", err)
                }
                m.state.queueMessageDelete.selected = 0
                return m.QueueDetailsSwitchPage(msg)
            }
        case key.Matches(msg, m.keys.Quit):
            m.state.queueMessageDelete.selected = 0
            return m.QueueDetailsSwitchPage(msg)
        }
    }

    return m, cmd
}
