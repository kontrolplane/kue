package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueDetails = "queue details"

func (m model) QueueDetailsView() string {
	return ""
}

func (m model) QueueDetailsSwitch(msg tea.Msg) (tea.Model, tea.Cmd) {
	m = m.SwitchPage(queueDetails)
	return m, nil
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	m = m.SwitchPage(queueDetails)

	return m, nil
}
