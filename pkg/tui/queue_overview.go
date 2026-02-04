package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

type queueOverviewState struct {
	selected int
	queues   []kue.Queue
	table    table.Model
}

var columnMap = map[int]string{
	0: "queue name",
	1: "last modified",
	2: "available",
	3: "not visible",
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
	// Clear any previous error
	m.error = ""

	// Switch page and trigger async load
	m = m.SwitchPage(queueOverview)
	m.loading = true
	m.loadingMsg = "Loading queues..."

	return m, commands.LoadQueues(m.context, m.client)
}

func (m model) NoQueuesFound() bool {
	return m.QueuesCount() == 0
}

func (m model) QueuesCount() int {
	return len(m.state.queueOverview.queues)
}

// initQueueOverviewTable initializes the queue overview table.
func initQueueOverviewTable(height int) table.Model {
	if height < minTableHeight {
		height = minTableHeight
	}

	t := table.New(
		table.WithColumns(queueOverviewColumns),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	t.SetStyles(styles.TableStyles())

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
			if len(m.state.queueOverview.queues) > 0 {
				selected := m.state.queueOverview.selected
				m.state.queueDetails.queue = m.state.queueOverview.queues[selected]
				return m.QueueDetailsSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Create):
			return m.QueueCreateSwitchPage(msg)
		case key.Matches(msg, m.keys.Delete):
			if len(m.state.queueOverview.queues) > 0 {
				selected := m.state.queueOverview.selected
				m.state.queueDelete.queue = m.state.queueOverview.queues[selected]
				return m.QueueDeleteSwitchPage(msg)
			}
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
	m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)

	if m.NoQueuesFound() {
		emptyMsg := lipgloss.NewStyle().
			Foreground(styles.MediumGray).
			Render("No queues found. Press Ctrl+N to create a new queue.")

		return lipgloss.Place(contentWidth, contentHeight-2,
			lipgloss.Center, lipgloss.Center,
			emptyMsg)
	}

	return m.state.queueOverview.table.View()
}
