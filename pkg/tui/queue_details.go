package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueDetailsState struct {
	selected int
	queue    kue.Queue
	messages []kue.Message
}

func (m model) QueueDetailsSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueDetails), nil
}

func (m model) nextMessage() (model, tea.Cmd) {
	n := m.state.queueDetails.selected + 1
	l := len(m.state.queueDetails.messages) - 1

	if n > l {
		n = l
	}

	m.state.queueDetails.selected = n
	return m, nil
}

func (m model) previousMessage() (model, tea.Cmd) {
	n := m.state.queueDetails.selected - 1

	if n < 0 {
		n = 0
	}

	m.state.queueDetails.selected = n
	return m, nil
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) QueueDetailsView() string {
	return ""
}
