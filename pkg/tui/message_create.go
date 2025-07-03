package tui

import (
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/huh"

    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueMessageCreateInput struct {
	messageBody   string
	deliveryDelay string
}

type queueMessageCreateState struct {
    queue kue.Queue // target queue selected from details view
    input queueMessageCreateInput
    form  huh.Form
    progress int // number of messages sent
    total int
    ch <-chan int
}

func (m model) QueueMessageCreateSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.state.queueMessageCreate.progress = 0
    m.state.queueMessageCreate.total = 0
    m.state.queueMessageCreate.form = huh.NewForm(
        huh.NewGroup(
            huh.NewText().Title("Message body (use '---' on a line by itself to separate multiple messages)").Height(8).
                Value(&m.state.queueMessageCreate.input.messageBody),
        ),
        huh.NewGroup(
            huh.NewText().Title("Delivery delay seconds (optional)").Value(&m.state.queueMessageCreate.input.deliveryDelay),
        ),
    )
    return m.SwitchPage(queueMessageCreate), nil
}

func (m model) QueueMessageCreateView() string {
    if m.state.queueMessageCreate.total > 0 {
        return fmt.Sprintf("Sending messages to queue %s: %d/%d sent", m.state.queueMessageCreate.queue.Name, m.state.queueMessageCreate.progress, m.state.queueMessageCreate.total)
    }
    return m.state.queueMessageCreate.form.View()
}

func (m model) QueueMessageCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.state.queueMessageCreate.total > 0 {
        switch t := msg.(type) {
        case progressMsg:
            if t.done {
                // Refresh messages list after sending
                return m.QueueDetailsSwitchPage(msg)
            }
            if t.sent > 0 {
                m.state.queueMessageCreate.progress = t.sent
            }
            return m, waitProgress(m.state.queueMessageCreate.ch)
        }
        return m, nil
    }

    var cmd tea.Cmd
    m.state.queueMessageCreate.form, cmd = m.state.queueMessageCreate.form.Update(msg)

    if m.state.queueMessageCreate.form.IsSubmitted() {
        bodies := kue.SplitBodies(m.state.queueMessageCreate.input.messageBody, kue.DefaultDelimiter)
        m.state.queueMessageCreate.total = len(bodies)
        delay := parseInt32(m.state.queueMessageCreate.input.deliveryDelay)
        progressCh := kue.SendMessages(m.client, m.context, m.state.queueMessageCreate.queue.Url, bodies, delay)
        m.state.queueMessageCreate.ch = progressCh
        return m, waitProgress(progressCh)
    }

    return m, cmd
}
