package tui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueMessageDeleteState keeps context about the message being deleted.
// The modal only needs to know which message is selected so that it can
// render a preview and issue the DeleteMessage API call once the user
// confirms.
//
// Unlike the queue-delete modal which offers two selectable buttons, this
// modal requires the user to explicitly type the lowercase character "y"
// to confirm the deletion. Therefore we only track whether the operation
// has already been executed.
// If aborted we leave the model untouched.
//
// selected is not required – we include a *deleted* bool so that we do not
// attempt to delete twice in case multiple key events arrive before the
// view is switched.

type queueMessageDeleteState struct {
    message kue.Message
    deleted bool
}

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    return m.SwitchPage(queueMessageDelete), nil
}

func (m model) QueueMessageDeleteView() string {
    if (m.state.queueMessageDelete == queueMessageDeleteState{}) {
        return ""
    }

    // Prepare excerpt – first 240 characters, replacing newlines with spaces.
    body := strings.ReplaceAll(m.state.queueMessageDelete.message.Body, "\n", " ")
    if len(body) > 240 {
        body = body[:240] + "…"
    }

    dialog := lipgloss.NewStyle().Padding(1, 3)
    promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("240")).Bold(true)

    content := fmt.Sprintf("%s\n\n%s\n\n%s", promptStyle.Render("Delete this message? Type 'y' to confirm:"), body, "Press any other key to cancel")

    return dialog.Render(content)
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.Type == tea.KeyRunes {
            if string(msg.Runes) == "y" && !m.state.queueMessageDelete.deleted {
                // Execute deletion
                err := kue.DeleteMessage(m.client, m.context, m.state.queueDetails.queue.Url, m.state.queueMessageDelete.message.ReceiptHandle)
                if err != nil {
                    m.error = fmt.Sprintf("Error deleting message: %v", err)
                }
                m.state.queueMessageDelete.deleted = true
            }
        }
        // For any key (including y) we return to details view afterwards
        return m.QueueDetailsSwitchPage(msg)
    default:
        return m, nil
    }
}
