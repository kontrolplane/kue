package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueDelete = "queue delete"

func (m model) QueueDeleteView() string {
	return ""
}

func (m model) QueueDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
