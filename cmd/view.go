package cmd

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var sqsOverviewStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

const padding = 40

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m model) View() string {

	var h string = "northernlights/kue"
	var f string = m.help.View(m.keys)

	var sqsOverviewColumns []table.Column = []table.Column{
		{
			Title: "account", Width: int(0.2 * float64(m.width-padding)),
		},
		{
			Title: "service endpoint", Width: int(0.5 * float64(m.width-padding)),
		},
		{
			Title: "queue name", Width: int(0.3 * float64(m.width-padding)),
		},
	}

	var sqsOverviewRows []table.Row = []table.Row{}

	for _, q := range m.queues {
		sqsOverviewRows = append(sqsOverviewRows, table.Row{
			q.AccountIdentifier,
			q.ServiceEndpoint,
			q.Name,
		})
	}

	t := table.New(
		table.WithColumns(sqsOverviewColumns),
		table.WithRows(sqsOverviewRows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#628049")).
		Bold(false)

	t.SetStyles(s)

	t.SetCursor(m.cursor)

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(h + "\n\n" + sqsOverviewStyle.Render(t.View()) + "\n\n" + f)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}
