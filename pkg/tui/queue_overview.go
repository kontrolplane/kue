package tui

import (
	"fmt"
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

func (m model) QueueOverviewSwitchPage(msg tea.Msg) (model, tea.Cmd) {

	log.Println("[QueueOverviewSwitchPage]")

	queues, err := kue.ListQueuesUrls(m.client, m.context)
	if err != nil {
		m.error = fmt.Sprintf("Error queue(s): %v", err)
	}

	m.state.queueOverview.selected = 0
	m.state.queueOverview.queues = queues

	m = m.SwitchPage(queueOverview)
	return m, updateQueuesCmd(queues)
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
	if m.state.queueOverview.selected < len(m.state.queueOverview.queues)-1 {
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

type UpdateQueuesMsg struct {
	Queues []kue.Queue
}

func updateQueuesCmd(queues []kue.Queue) tea.Cmd {
	return func() tea.Msg {
		return UpdateQueuesMsg{Queues: queues}
	}
}

func (m model) QueueOverviewUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down):
			m, cmd = m.nextQueue()
		case key.Matches(msg, m.keys.Up):
			m, cmd = m.previousQueue()
		case key.Matches(msg, m.keys.View):
			selected := m.state.queueOverview.selected
			m.state.queueDetails.queue = m.state.queueOverview.queues[selected]
			return m.QueueDetailsSwitchPage(msg)
		case key.Matches(msg, m.keys.Create):
			return m.QueueCreateSwitchPage(msg)
		case key.Matches(msg, m.keys.Delete):
			selected := m.state.queueOverview.selected
			m.state.queueDelete.queue = m.state.queueOverview.queues[selected]
			return m.QueueDeleteSwitchPage(msg)
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		default:
			m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
		}
	default:
		m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
	}

	return m, cmd
}

func (m model) QueueOverviewView() string {

	if m.NoQueuesFound() {
		m.error = fmt.Sprint("No queues found")
	}

	m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)

	return m.state.queueOverview.table.View()
}
