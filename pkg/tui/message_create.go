package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameMessageCreation = "message creation"

func (m model) MessageCreationView() string {
	return ""
}

func (m model) MessageCreationUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
