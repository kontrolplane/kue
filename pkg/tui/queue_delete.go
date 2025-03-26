package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueDelete = "queue delete"

type queueDeleteState struct {
}

func (m model) QueueDeleteSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueDelete), nil
}

func (m model) QueueDeleteView() string {
	return ""
}

func (m model) QueueDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
