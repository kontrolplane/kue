package tui

import (
    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/lipgloss"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/kontrolplane/kue/pkg/tui/commands"
)

type queueCreateInput struct {
	queueType              string
	name                   string
	region                 string
	visibilityTimeout      int
	visibilityTimeoutType  string
	messageRetention       int
	messageRetentionType   string
	deliveryDelay          int
	deliveryDelayType      string
	maximumMessageSize     int
	receiveMessageWaitTime int
}

type queueCreateState struct {
	input queueCreateInput
	form  huh.Form
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
    // temporary simple view: instruct form not implemented
    return lipgloss.NewStyle().Render("[queue create form placeholder]")
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    // For MVP, when user presses Enter (SubmitMsg from huh), trigger create queue
    var cmd tea.Cmd
    m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)

    switch msg.(type) {
    case huh.SubmitMsg:
        // TODO: gather data from form; using placeholder values
        name := m.state.queueCreate.input.name
        if name == "" {
            name = "sample-queue" // placeholder default
        }
        m.creatingQueue = true
        m.spinner = spinner.New()
        return m.SwitchPage(queueOverview), tea.Batch(cmd, commands.CreateQueueCmd(m.client, m.context, name, map[string]string{}, nil))
    }

    return m, cmd
}
