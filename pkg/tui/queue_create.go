package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueCreation = "queue creation"

func (m model) QueueCreationView() string {
	return ""
}

func (m model) QueueCreationUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
