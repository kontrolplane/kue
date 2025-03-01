package cmd

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var sqsOverviewStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var sqsOverviewColumns []table.Column = []table.Column{
	{
		Title: "Type", Width: 10,
	},
	{
		Title: "Service Endpoint", Width: 50,
	},
	{
		Title: "Account Identifier", Width: 30,
	},
	{
		Title: "Queue Name", Width: 30,
	},
}

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m model) View() string {

	var h string = "northernlights/kue"
	var f string = m.help.View(m.keys)

	var sqsOverviewRows []table.Row = []table.Row{}

	for _, q := range m.queues {
		sqsOverviewRows = append(sqsOverviewRows, table.Row{
			q.Protocol,
			q.ServiceEndpoint,
			q.AccountIdentifier,
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
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t.SetStyles(s)

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(h + "\n\n" + sqsOverviewStyle.Render(t.View()) + "\n\n" + f)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}
