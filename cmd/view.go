package cmd

import (
	"github.com/charmbracelet/lipgloss"
)

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m model) View() string {

	var h string = "northernlights/kue"
	var c string = m.queues[0]
	var f string = m.help.View(m.keys)

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(h + "\n\n" + c + "\n\n" + f)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}
