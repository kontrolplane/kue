package cmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
)

var sqsOverviewStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// View renders the program's UI, which is just a string. The view is
// rendered after every Update.
func (m model) View() string {

	var h string = fmt.Sprintf("%s/%s", ProjectName, ProgramName)
	var f string = m.help.View(m.keys)

	var sqsOverviewColumns []table.Column = []table.Column{
		{
			Title: "account", Width: 20,
		},
		{
			Title: "service endpoint", Width: 50,
		},
		{
			Title: "queue name", Width: 40,
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
		table.WithHeight(10),
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
