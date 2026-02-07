package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueOverviewState holds the state for the queue listing view.
type queueOverviewState struct {
	selected      int
	queues        []kue.Queue
	table         table.Model
	selectedItems map[int]bool // tracks which items are selected for bulk operations
}

// Queue table column definitions.
var columnMap = map[int]string{
	0: "queue name",
	1: "type",
	2: "available",
	3: "not visible",
	4: "delayed",
	5: "visibility",
	6: "retention",
	7: "last updated",
}

var queueOverviewColumns []table.Column = []table.Column{
	{
		Title: columnMap[0], Width: 40,
	},
	{
		Title: columnMap[1], Width: 10,
	},
	{
		Title: columnMap[2], Width: 10,
	},
	{
		Title: columnMap[3], Width: 10,
	},
	{
		Title: columnMap[4], Width: 10,
	},
	{
		Title: columnMap[5], Width: 10,
	},
	{
		Title: columnMap[6], Width: 10,
	},
	{
		Title: columnMap[7], Width: 20,
	},
}

func (m model) QueueOverviewSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	m = m.SwitchPage(queueOverview)
	m.loading = true
	m.loadingMsg = "Loading queues..."
	m.state.queueOverview.selectedItems = make(map[int]bool)
	return m, commands.LoadQueues(m.context, m.client)
}

func (m model) NoQueuesFound() bool {
	return m.QueuesCount() == 0
}

func (m model) QueuesCount() int {
	return len(m.state.queueOverview.queues)
}

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

func (m model) toggleQueueSelection() (model, tea.Cmd) {
	if len(m.state.queueOverview.queues) == 0 {
		return m, nil
	}
	idx := m.state.queueOverview.selected
	if m.state.queueOverview.selectedItems == nil {
		m.state.queueOverview.selectedItems = make(map[int]bool)
	}
	if m.state.queueOverview.selectedItems[idx] {
		delete(m.state.queueOverview.selectedItems, idx)
	} else {
		m.state.queueOverview.selectedItems[idx] = true
	}
	return m, nil
}

func (m model) getSelectedQueues() []kue.Queue {
	var queues []kue.Queue
	for idx := range m.state.queueOverview.selectedItems {
		if idx < len(m.state.queueOverview.queues) {
			queues = append(queues, m.state.queueOverview.queues[idx])
		}
	}
	return queues
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
		case key.Matches(msg, m.keys.Select):
			m, cmd = m.toggleQueueSelection()
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
				// If items are selected, delete selected items; otherwise delete current item
				if len(m.state.queueOverview.selectedItems) > 0 {
					m.state.queueDelete.queues = m.getSelectedQueues()
				} else {
					selected := m.state.queueOverview.selected
					m.state.queueDelete.queues = []kue.Queue{m.state.queueOverview.queues[selected]}
				}
				return m.QueueDeleteSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Quit):
			// If items are selected, clear selection instead of quitting
			if len(m.state.queueOverview.selectedItems) > 0 {
				m.state.queueOverview.selectedItems = make(map[int]bool)
				return m, nil
			}
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
	// Rebuild table rows to reflect current selection state
	m = m.updateQueueOverviewTable()
	tableView := m.state.queueOverview.table.View()

	if m.NoQueuesFound() {
		emptyMsg := lipgloss.NewStyle().
			Foreground(styles.MediumGray).
			Render("No queues found. Press Ctrl+N to create a new queue.")

		return tableView + "\n\n" + emptyMsg
	}

	// Show selection count if items are selected
	if len(m.state.queueOverview.selectedItems) > 0 {
		selectionInfo := lipgloss.NewStyle().
			Foreground(styles.AccentColor).
			Render(fmt.Sprintf("%d queue(s) selected - press ctrl + d to delete", len(m.state.queueOverview.selectedItems)))
		return tableView + "\n" + selectionInfo
	}

	return tableView
}
