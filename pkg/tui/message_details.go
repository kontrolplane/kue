package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameMessageDetails = "message details"

func (m model) MessageDetailsView() string {
	return ""
}

func (m model) MessageDetailsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
