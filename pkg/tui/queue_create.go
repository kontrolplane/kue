package tui

import (
    "fmt"
    "github.com/charmbracelet/huh"

    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
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
	form  *huh.Form
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
    // Render the Huh form view when initialised; if not, return placeholder
    if m.state.queueCreate.form == nil {
        return "Loading form..."
    }
    return m.state.queueCreate.form.View()
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    // Lazily build the form the first time we enter the page
    if m.state.queueCreate.form == nil {
        form := huh.NewForm(huh.NewGroup(
            huh.NewSelect[string]().Title("Queue type").Options(
                huh.NewOption("Standard", "standard"),
                huh.NewOption("FIFO", "fifo"),
            ).Value(&m.state.queueCreate.input.queueType),
            huh.NewInput().Title("Name").CharLimit(80).Value(&m.state.queueCreate.input.name),
        )).WithTheme(huh.ThemeBase())

        m.state.queueCreate.form = form
    }

    if m.page == queueCreate {
        m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)
    }

    // When the form is submitted we actually call AWS and then go back to overview
    if m.state.queueCreate.form != nil && m.state.queueCreate.form.State == huh.StateCompleted {
        fifo := m.state.queueCreate.input.queueType == "fifo"
        if err := kue.CreateQueue(m.client, m.context, m.state.queueCreate.input.name, fifo); err != nil {
            m.error = fmt.Sprintf("Error creating queue: %v", err)
        }
        // refresh list
        return m.QueueOverviewSwitchPage(msg)
    }

    return m, cmd
}
