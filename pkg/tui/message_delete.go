package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameMessageDelete = "message delete"

func (m model) MessageDeleteView() string {
	return ""
}

func (m model) MessageDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
