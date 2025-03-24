package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueMessageDelete = "queue message delete"

func (m model) QueueMessageDeleteSwitch(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueMessageDelete), nil
}

func (m model) QueueMessageDeleteView() string {
	return ""
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
