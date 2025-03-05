package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

var (
	ProjectName = "kontrolplane"
	ProgramName = "kue"
)

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle(ProgramName)
}

func Execute() {

	// Context is a context that can be used to cancel the program
	ctx := context.Background()

	// Get the list of queues
	queues, err := ListQueues(ctx)

	var queueOverviewColumns []table.Column = []table.Column{
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

	var queueOverviewRows []table.Row = []table.Row{}

	for _, q := range queues {
		queueOverviewRows = append(queueOverviewRows, table.Row{
			q.AccountIdentifier,
			q.ServiceEndpoint,
			q.Name,
		})
	}

	t := table.New(
		table.WithColumns(queueOverviewColumns),
		table.WithRows(queueOverviewRows),
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

	f, err := tea.LogToFile("debug.log", "help")
	if err != nil {
		fmt.Println("Couldn't open a file for logging:", err)
		os.Exit(1)
	}
	defer f.Close() // nolint:errcheck

	// Run the program
	if _, err := tea.NewProgram(model{
		cursor:   0,
		keys:     keys,
		help:     help.New(),
		selected: make(map[int]struct{}),

		table: t,
	}, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
