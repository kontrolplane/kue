package tui

import (
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
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
	3: "in flight",
	4: "delayed",
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
		Title: columnMap[3], Width: 15,
	},
	{
		Title: columnMap[4], Width: 10,
	},
}

func (m model) QueueOverviewSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueOverview), nil
}

func (m model) NoQueuesFound() bool {
	return m.QueuesCount() == 0
}

func (m model) QueuesCount() int {
	return len(m.state.queueOverview.queues)
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

func (m model) nextQueue() (model, tea.Cmd) {
    if m.state.queueOverview.selected < len(m.state.queueOverview.queues) - 1 {
        m.state.queueOverview.selected++
    }
    return m, nil
}

func (m model) previousQueue() (model, tea.Cmd) {
    if m.state.queueOverview.selected > 0 {
        m.state.queueOverview.selected--
    }
    return m, nil
}

func (m model) QueueOverviewUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down):
			m, cmd = m.nextQueue()
			m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)
		case key.Matches(msg, m.keys.Up):
			m, cmd = m.previousQueue()
			m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)
		case key.Matches(msg, m.keys.View):
			selected := m.state.queueOverview.table.Cursor()
			if selected >= 0 && selected < len(m.state.queueOverview.queues) {
				m.previous = m.page
				m.page = queueDetails
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

func (m model) QueueOverviewView() string {

	log.Println("[QueueOverviewView] queues:", m.state.queueOverview.queues)

	if m.NoQueuesFound() {
		return "No queues found."
	}

	var queueOverviewRows []table.Row

	for _, queue := range m.state.queueOverview.queues {
		queueOverviewRows = append(queueOverviewRows, table.Row{
			queue.Name,
			queue.LastModified,
			queue.ApproximateNumberOfMessages,
			queue.ApproximateNumberOfMessagesNotVisible,
			queue.ApproximateNumberOfMessagesDelayed,
		})
	}

	m.state.queueOverview.table.SetRows(queueOverviewRows)
	m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)

	return m.state.queueOverview.table.View()
}
