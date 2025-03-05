package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var queueOverviewStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m model) View() string {
	var h string = fmt.Sprintf("%s/%s • %s", ProjectName, ProgramName, "queue overview")
	var f string = m.help.View(m.keys)

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(h + "\n\n" + queueOverviewStyle.Render(m.table.View()) + "\n\n" + f)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}
