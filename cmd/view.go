package cmd

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m model) View() string {

	var c string

	content := lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(c)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}
