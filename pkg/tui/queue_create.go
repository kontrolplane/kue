package tui

import (
	"github.com/charmbracelet/huh"

	tea "github.com/charmbracelet/bubbletea"
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
	formView := m.state.queueCreate.form.View()
	if formView == "" {
		return "Queue Creation (in progress)\n(Not yet implemented)"
	}
	return formView
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	// pass events to the form if available
	var cmd tea.Cmd
	form := &m.state.queueCreate.form
	if form != nil {
		*form, cmd = form.Update(msg)
	}
	// When the form is submitted, gather input and create the queue
	if form != nil && form.Submitted() {
		// Here we would extract the fields and call the backend create logic
		// Reset/close the form for demonstration purposes
		m.error = "Queue creation submitted (not fully implemented yet)"
	}
	return m, cmd
}
