package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

var (
	viewNameQueueOverview = "queue overview"
	errNoQueuesFound      = "No queues found"
)

type queueOverviewState struct {
	selected int
	queues   []kue.Queue
	table    table.Model
}

var columnMap = map[int]string{
	0: "queue name",
	1: "last modified",
	2: "messages",
	3: "delayed",
	4: "not visible",
}

var queueOverviewColumns []table.Column = []table.Column{
	{
		Title: columnMap[0], Width: 40,
	},
	{
		Title: columnMap[1], Width: 20,
	},
	{
		Title: columnMap[2], Width: 10,
	},
	{
		Title: columnMap[3], Width: 10,
	},
	{
		Title: columnMap[4], Width: 15,
	},
}

func (m model) QueueOverviewSwitch(msg tea.Msg) (tea.Model, tea.Cmd) {
	m = m.SwitchPage(queueOverview)
	return m, nil
}

// Helper function to initialize the queue overview table
func initQueueOverviewTable() table.Model {

	t := table.New(
		table.WithColumns(queueOverviewColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
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

	log.Println("[QueueOverviewView] queues:", m.state.queueOverview.queues)

	var queueOverviewRows []table.Row

	for _, q := range m.state.queueOverview.queues {
		queueOverviewRows = append(queueOverviewRows, table.Row{
			q.Name,
			q.LastModified,
			q.ApproximateNumberOfMessages,
			q.ApproximateNumberOfMessagesDelayed,
			q.ApproximateNumberOfMessagesNotVisible,
		})
	}

	log.Println("[QueueOverviewView] table rows:", queueOverviewRows)

	m.state.queueOverview.table.SetRows(queueOverviewRows)

	return m.state.queueOverview.table.View()
}

func (m model) QueueOverviewUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
		case key.Matches(msg, m.keys.Down):
			m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)

		case key.Matches(msg, m.keys.View):
			selected := m.state.queueOverview.table.Cursor()
			if selected >= 0 && selected < len(m.state.queueOverview.queues) {
				m.previous = m.page
				m.page = queueDetails
				m.viewName = viewNameQueueDetails
				m.state.queueDetails.queue = m.state.queueOverview.queues[selected]
			}

		default:
			m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
		}

	default:
		m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
	}

	return m, cmd
}
