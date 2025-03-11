package tui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	sqs "github.com/kontrolplane/kue/pkg/sqs"
)

var (
	viewNameQueueOverview = "queue overview"
	errNoQueuesFound      = "No queues found"
)

type queueOverviewState struct {
	selected int
	queues   []sqs.Queue
	table    table.Model
}

var columnMap = map[int]string{
	0: "account",
	1: "service endpoint",
	2: "queue name",
}

var queueOverviewColumns []table.Column = []table.Column{
	{
		Title: columnMap[0], Width: 20,
	},
	{
		Title: columnMap[1], Width: 50,
	},
	{
		Title: columnMap[2], Width: 40,
	},
}

var queueOverviewRows []table.Row = []table.Row{
	{"account-1", "https://service-endpoint-1", "queue-1"},
	{"account-2", "https://service-endpoint-2", "queue-2"},
	{"account-3", "https://service-endpoint-3", "queue-3"},
	{"account-4", "https://service-endpoint-4", "queue-4"},
}

func (m model) QueueOverviewSwitch(msg tea.Msg) (tea.Model, tea.Cmd) {
	m = m.SwitchPage(queueOverview)
	return m, nil
}

// Helper function to initialize the queue overview table
func initQueueOverviewTable() table.Model {

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

	return t
}

func (m model) QueueOverviewView() string {

	/**
	for _, q := range queues {
		queueOverviewRows = append(queueOverviewRows, table.Row{
			q.AccountIdentifier,
			q.ServiceEndpoint,
			q.Name,
		})
	}
	*/

	return m.state.queueOverview.table.View()
}

func (m model) QueueOverviewUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
