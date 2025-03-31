package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) QueueMessageDetailsSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueMessageDetails), nil
}

func (m model) QueueMessageDetailsView() string {
	return ""
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
