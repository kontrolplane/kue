package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
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
	filtering     bool
	filterInput   textinput.Model
	filterText    string
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
	m.state.queueOverview.filtering = false
	m.state.queueOverview.filterText = ""
	m.state.queueOverview.filterInput = initFilterInput()
	return m, commands.LoadQueues(m.context, m.client)
}

func initFilterInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Type to filter..."
	ti.CharLimit = 50
	ti.Width = 30
	return ti
}

func (m model) getFilteredQueues() []kue.Queue {
	if m.state.queueOverview.filterText == "" {
		return m.state.queueOverview.queues
	}
	filter := strings.ToLower(m.state.queueOverview.filterText)
	var filtered []kue.Queue
	for _, q := range m.state.queueOverview.queues {
		if strings.Contains(strings.ToLower(q.Name), filter) {
			filtered = append(filtered, q)
		}
	}
	return filtered
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
	filteredQueues := m.getFilteredQueues()
	if m.state.queueOverview.selected < len(filteredQueues)-1 {
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

	// Handle filter mode
	if m.state.queueOverview.filtering {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEsc:
				m.state.queueOverview.filtering = false
				m.state.queueOverview.filterInput.Blur()
				return m, nil
			case tea.KeyEnter:
				m.state.queueOverview.filtering = false
				m.state.queueOverview.filterText = m.state.queueOverview.filterInput.Value()
				m.state.queueOverview.filterInput.Blur()
				m.state.queueOverview.selected = 0
				return m, nil
			}
		}
		m.state.queueOverview.filterInput, cmd = m.state.queueOverview.filterInput.Update(msg)
		// Live filtering as user types
		m.state.queueOverview.filterText = m.state.queueOverview.filterInput.Value()
		m.state.queueOverview.selected = 0
		return m, cmd
	}

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Filter):
			m.state.queueOverview.filtering = true
			m.state.queueOverview.filterInput.Focus()
			return m, textinput.Blink
		case key.Matches(msg, m.keys.Down):
			m, cmd = m.nextQueue()
		case key.Matches(msg, m.keys.Up):
			m, cmd = m.previousQueue()
		case key.Matches(msg, m.keys.Select):
			m, cmd = m.toggleQueueSelection()
		case key.Matches(msg, m.keys.View):
			filteredQueues := m.getFilteredQueues()
			if len(filteredQueues) > 0 {
				selected := m.state.queueOverview.selected
				m.state.queueDetails.queue = filteredQueues[selected]
				return m.QueueDetailsSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Create):
			return m.QueueCreateSwitchPage(msg)
		case key.Matches(msg, m.keys.Delete):
			filteredQueues := m.getFilteredQueues()
			if len(filteredQueues) > 0 {
				// If items are selected, delete selected items; otherwise delete current item
				if len(m.state.queueOverview.selectedItems) > 0 {
					m.state.queueDelete.queues = m.getSelectedQueues()
				} else {
					selected := m.state.queueOverview.selected
					m.state.queueDelete.queues = []kue.Queue{filteredQueues[selected]}
				}
				return m.QueueDeleteSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Quit):
			// If filtering, clear filter
			if m.state.queueOverview.filterText != "" {
				m.state.queueOverview.filterText = ""
				m.state.queueOverview.filterInput.SetValue("")
				m.state.queueOverview.selected = 0
				return m, nil
			}
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
	m = m.updateQueueOverviewTableFiltered()
	tableView := m.state.queueOverview.table.View()

	filteredQueues := m.getFilteredQueues()
	if len(filteredQueues) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(styles.MediumGray).
			Render("No queues found. Press Ctrl+N to create a new queue.")

		return tableView + "\n\n" + emptyMsg
	}

	return tableView
}
