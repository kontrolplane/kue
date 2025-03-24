package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueCreate = "queue create"

func (m model) QueueCreateSwitch(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
	return ""
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
