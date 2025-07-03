package tui

import (
    "fmt"
    "log"

    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/lipgloss"

    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueMessageRequeueState struct {
    message  kue.Message
    queueUrl string
    selected int // 0 – abort, 1 – confirm
}

func (m model) QueueMessageRequeueSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    log.Println("[QueueMessageRequeueSwitchPage]")
    return m.SwitchPage(queueMessageRequeue), nil
}

func (m model) QueueMessageRequeueView() string {
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

    confirm := "yes"
    abort := "no"

    if m.state.queueMessageRequeue.selected == 0 {
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

    d := lipgloss.JoinVertical(
        lipgloss.Center,
        "re-queue message",
        "",
        "are you sure you want to return the message to the queue?",
        "",
        buttons,
    )

    return dialog.Render(d)
}

func (m model) switchRequeueOption() (model, tea.Cmd) {
    m.state.queueMessageRequeue.selected = (m.state.queueMessageRequeue.selected + 1) % 2
    return m, nil
}

func (m model) QueueMessageRequeueUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Left):
            m, cmd = m.switchRequeueOption()
        case key.Matches(msg, m.keys.Right):
            m, cmd = m.switchRequeueOption()
        case key.Matches(msg, m.keys.View):
            if m.state.queueMessageRequeue.selected == 0 {
                // abort
                return m.QueueDetailsSwitchPage(msg)
            }
            // confirm requeue (visibility timeout = 0)
            err := kue.ChangeMessageVisibility(
                m.client,
                m.context,
                m.state.queueDetails.queue.Url,
                m.state.queueMessageRequeue.message.ReceiptHandle,
                0,
            )
            if err != nil {
                m.error = fmt.Sprintf("Error re-queuing message: %v", err)
            } else {
                // After requeue, fetch fresh messages to update view
                newMsgs, err := kue.FetchQueueMessages(m.client, m.context, m.state.queueDetails.queue.Url, 10)
                if err == nil {
                    m.state.queueDetails.messages = newMsgs
                }
            }
            return m.QueueDetailsSwitchPage(msg)
        case key.Matches(msg, m.keys.Quit):
            return m.QueueDetailsSwitchPage(msg)
        }
    }
    return m, cmd
}
