package tui

import (
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/textarea"
    "github.com/charmbracelet/lipgloss"
    "github.com/charmbracelet/bubbles/key"
)

type queueMessageCreateInput struct {
	messageBody   string
	deliveryDelay string
}

// queueMessageCreateState keeps UI state for composing a new message.
type queueMessageCreateState struct {
    textarea textarea.Model // textarea for editing the message body
}

func (m model) QueueMessageCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    // Initialize textarea when entering this page.
    ta := textarea.New()
    ta.Placeholder = "Enter message body..."
    ta.Focus()
    ta.Prompt = ""
    ta.CharLimit = 0 // unlimited, we'll enforce via BytesRemaining
    ta.SetWidth(60)
    ta.SetHeight(10)

    m.state.queueMessageCreate = queueMessageCreateState{
        textarea: ta,
    }

    return m.SwitchPage(queueMessageCreate), nil
}

func (m model) QueueMessageCreateView() string {
    taView := m.state.queueMessageCreate.textarea.View()

    remaining := BytesRemaining(m.state.queueMessageCreate.textarea.Value())
    var remainingStyle lipgloss.Style
    if remaining < 0 {
        remainingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // red
    } else {
        remainingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
    }

    counter := fmt.Sprintf("%d B remaining (max %d)", remaining, MaxMessageBytes)

    counterView := remainingStyle.Render(counter)

    return taView + "\n" + counterView
}

func (m model) QueueMessageCreateUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.View): // enter key tries to send
            body := m.state.queueMessageCreate.textarea.Value()
            if BytesRemaining(body) < 0 {
                // exceed limit, set error and stay
                m.error = "Message exceeds maximum size of 256KB"
                return m, nil
            }
            // TODO: implement sending logic
            // For now, just go back to details or overview
            return m.QueueOverviewSwitchPage(msg)
        case key.Matches(msg, m.keys.Quit):
            return m.QueueOverviewSwitchPage(msg)
        }
    }

    m.state.queueMessageCreate.textarea, cmd = m.state.queueMessageCreate.textarea.Update(msg)
    return m, cmd
}
