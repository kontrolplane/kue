package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueMessageDelete), nil
}

func (m model) QueueMessageDeleteView() string {
	return ""
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
