package tui

import (
    "context"
    "fmt"

    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/lipgloss"
    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueMessageDeleteState keeps UI state for deleting a set of messages
// Phases: confirm -> progress -> done
// selected==0 means "no" (abort) button selected during confirm phase, 1==yes
// During progress phase, statuses slice mirrors messages slice and is updated
// when we receive MessageDeletedMsg.

type queueMessageDeleteState struct {
    queue        kue.Queue
    messages     []kue.Message
    selected     int      // button selection for confirm dialog
    phase        string   // "confirm"|"progress"|"done"
    statuses     []string // per message status: "pending","ok","err: ..."
    completedCnt int
}

// MessageDeletedMsg is sent when a single message deletion attempt finishes.
// Index references messages slice index.

type MessageDeletedMsg struct {
    Index int
    Err   error
}

func deleteMessageCmd(qURL string, receiptHandle string, idx int, client kue.SQSMessageDeleter, ctx context.Context) tea.Cmd {
    return func() tea.Msg {
        err := kue.DeleteMessage(client, ctx, qURL, receiptHandle)
        return MessageDeletedMsg{Index: idx, Err: err}
    }
}

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
    // when switching in, ensure state.phase = "confirm" and statuses prepared
    m.state.queueMessageDelete.phase = "confirm"
    m.state.queueMessageDelete.selected = 1 // default to "yes" for quick enter press
    m.state.queueMessageDelete.statuses = make([]string, len(m.state.queueMessageDelete.messages))
    for i := range m.state.queueMessageDelete.statuses {
        m.state.queueMessageDelete.statuses[i] = "pending"
    }
    m.state.queueMessageDelete.completedCnt = 0
    return m.SwitchPage(queueMessageDelete), nil
}

func (m model) QueueMessageDeleteView() string {
    qm := m.state.queueMessageDelete

    dialog := lipgloss.NewStyle().Padding(1, 3)
    secondary := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Background(lipgloss.Color("#628049")).Padding(0, 3)
    primary := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Background(lipgloss.Color("240")).Padding(0, 3)

    switch qm.phase {
    case "confirm":
        confirm := "yes"
        abort := "no"
        if qm.selected == 1 {
            confirm = primary.Render(confirm)
            abort = secondary.Render(abort)
        } else {
            confirm = secondary.Render(confirm)
            abort = primary.Render(abort)
        }
        buttons := lipgloss.JoinHorizontal(lipgloss.Center, abort, "    ", confirm)
        body := lipgloss.JoinVertical(lipgloss.Center,
            fmt.Sprintf("warning: delete %d messages", len(qm.messages)),
            "",
            "are you sure you want to delete the selected messages?",
            "",
            buttons,
        )
        return dialog.Render(body)

    case "progress":
        rows := make([]string, len(qm.messages))
        for i, msg := range qm.messages {
            status := qm.statuses[i]
            idPart := msg.MessageID
            if len(idPart) > 8 {
                idPart = idPart[:8]
            }
            rows[i] = fmt.Sprintf("%s ... %s", idPart, status)
        }
        body := lipgloss.JoinVertical(lipgloss.Left, append([]string{"deleting messages:"}, rows...)...)
        return dialog.Render(body)

    case "done":
        summary := fmt.Sprintf("Deleted %d/%d messages. Press enter to return.", qm.completedCnt, len(qm.messages))
        return dialog.Render(summary)
    default:
        return dialog.Render("unknown state")
    }
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmds []tea.Cmd

    qm := &m.state.queueMessageDelete

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch qm.phase {
        case "confirm":
            switch {
            case key.Matches(msg, m.keys.Left):
                qm.selected = (qm.selected + 1) % 2
            case key.Matches(msg, m.keys.Right):
                qm.selected = (qm.selected + 1) % 2
            case key.Matches(msg, m.keys.View): // Enter
                if qm.selected == 1 {
                    // proceed deletion
                    qm.phase = "progress"
                    for i, message := range qm.messages {
                        cmds = append(cmds, deleteMessageCmd(qm.queue.Url, message.ReceiptHandle, i, m.client, m.context))
                    }
                } else {
                    // abort
                    // clear selection in queue details
                    m.state.queueDetails.selectedRows = map[int]bool{}
                    return m.QueueDetailsSwitchPage(msg)
                }
            case key.Matches(msg, m.keys.Quit):
                // abort
                m.state.queueDetails.selectedRows = map[int]bool{}
                return m.QueueDetailsSwitchPage(msg)
            }
        case "done":
            switch {
            case key.Matches(msg, m.keys.View):
                // return
                m.state.queueDetails.selectedRows = map[int]bool{}
                return m.QueueDetailsSwitchPage(msg)
            case key.Matches(msg, m.keys.Quit):
                m.state.queueDetails.selectedRows = map[int]bool{}
                return m.QueueDetailsSwitchPage(msg)
            }
        }

    case MessageDeletedMsg:
        if qm.phase == "progress" {
            if msg.Err != nil {
                qm.statuses[msg.Index] = "err"
            } else {
                qm.statuses[msg.Index] = "ok"
            }
            qm.completedCnt++
            if qm.completedCnt == len(qm.messages) {
                qm.phase = "done"
            }
        }
    }

    // aggregate cmds if any
    var cmd tea.Cmd
    if len(cmds) == 1 {
        cmd = cmds[0]
    } else if len(cmds) > 1 {
        cmd = tea.Batch(cmds...)
    }
    return m, cmd
}
