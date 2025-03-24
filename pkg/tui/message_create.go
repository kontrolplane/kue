package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueMessageCreate = "queue message create"

func (m model) QueueMessageCreateSwitch(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueMessageCreate), nil
}

func (m model) QueueMessageCreateView() string {
	return ""
}

func (m model) QueueMessageCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
