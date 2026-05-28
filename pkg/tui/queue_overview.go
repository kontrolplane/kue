package tui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
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
	filter        filterModel
}

func queueOverviewColumns() []table.Column {
	const (
		numColumns       = 8
		cellPadding      = numColumns * 2 // bubbles table adds 1 char padding on each side per cell
		typeWidth        = 9
		availableWidth   = 10
		notVisibleWidth  = 12
		delayedWidth     = 9
		visibilityWidth  = 10
		retentionWidth   = 10
		lastUpdatedWidth = 24
	)
	fixedWidth := typeWidth + availableWidth + notVisibleWidth + delayedWidth + visibilityWidth + retentionWidth + lastUpdatedWidth
	nameWidth := contentWidth - fixedWidth - cellPadding

	return []table.Column{
		{Title: "Queue Name", Width: nameWidth},
		{Title: "Type", Width: typeWidth},
		{Title: "Available", Width: availableWidth},
		{Title: "Not Visible", Width: notVisibleWidth},
		{Title: "Delayed", Width: delayedWidth},
		{Title: "Visibility", Width: visibilityWidth},
		{Title: "Retention", Width: retentionWidth},
		{Title: "Last Updated", Width: lastUpdatedWidth},
	}
}

func (m model) QueueOverviewSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	m = m.SwitchPage(queueOverview)
	m.loading = true
	m.loadingMsg = "Loading queues..."
	m.state.queueOverview.selectedItems = make(map[int]bool)
	m.state.queueOverview.filter = newFilter("Type to filter...")
	return m, commands.LoadQueues(m.context, m.client)
}

func initQueueOverviewTable(height int) table.Model {
	if height < minTableHeight {
		height = minTableHeight
	}

	t := table.New(
		table.WithColumns(queueOverviewColumns()),
		table.WithFocused(true),
		table.WithWidth(contentWidth),
		table.WithHeight(height),
	)

	t.SetStyles(styles.TableStyles())

	return t
}

func (m model) getFilteredQueues() []kue.Queue {
	if m.state.queueOverview.filter.text == "" {
		return m.state.queueOverview.queues
	}
	var filtered []kue.Queue
	for _, q := range m.state.queueOverview.queues {
		if m.state.queueOverview.filter.Matches(q.Name) {
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

func (m model) nextQueue() model {
	filteredQueues := m.getFilteredQueues()
	if m.state.queueOverview.selected < len(filteredQueues)-1 {
		m.state.queueOverview.selected++
	}
	return m
}

func (m model) previousQueue() model {
	if m.state.queueOverview.selected > 0 {
		m.state.queueOverview.selected--
	}
	return m
}

func (m model) toggleQueueSelection() model {
	if len(m.state.queueOverview.queues) == 0 {
		return m
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
	return m
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
	// Handle filter mode - delegate to filter model
	if m.state.queueOverview.filter.active {
		var cmd tea.Cmd
		m.state.queueOverview.filter, cmd = m.state.queueOverview.filter.Update(msg)
		m.state.queueOverview.selected = 0
		m = m.rebuildQueueTable()
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Filter):
			var cmd tea.Cmd
			m.state.queueOverview.filter, cmd = m.state.queueOverview.filter.Activate()
			return m, cmd
		case key.Matches(msg, m.keys.Down):
			m = m.nextQueue()
			m = m.rebuildQueueTable()
		case key.Matches(msg, m.keys.Up):
			m = m.previousQueue()
			m = m.rebuildQueueTable()
		case key.Matches(msg, m.keys.Select):
			m = m.toggleQueueSelection()
			m = m.rebuildQueueTable()
		case key.Matches(msg, m.keys.View):
			filteredQueues := m.getFilteredQueues()
			if len(filteredQueues) > 0 {
				selected := m.state.queueOverview.selected
				m.state.queueDetails.queue = filteredQueues[selected]
				return m.QueueDetailsSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Create):
			return m.QueueCreateSwitchPage(msg)
		case key.Matches(msg, m.keys.Purge):
			filteredQueues := m.getFilteredQueues()
			if len(filteredQueues) > 0 {
				selected := m.state.queueOverview.selected
				m.state.queuePurge.queue = filteredQueues[selected]
				m.state.queuePurge.fromOverview = true
				return m.QueuePurgeSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Redrive):
			filteredQueues := m.getFilteredQueues()
			if len(filteredQueues) > 0 {
				selected := m.state.queueOverview.selected
				queue := filteredQueues[selected]
				sourceArn := m.findSourceQueueArn(queue.Arn)
				if sourceArn != "" {
					m.state.queueRedrive.queue = queue
					m.state.queueRedrive.destinationArn = sourceArn
					m.state.queueRedrive.fromOverview = true
					return m.QueueRedriveSwitchPage(msg)
				}
				m.error = "This queue is not a dead-letter queue"
			}
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
			if m.state.queueOverview.filter.text != "" {
				m.state.queueOverview.filter = m.state.queueOverview.filter.Clear()
				m.state.queueOverview.selected = 0
				m = m.rebuildQueueTable()
				return m, nil
			}
			// If items are selected, clear selection instead of quitting
			if len(m.state.queueOverview.selectedItems) > 0 {
				m.state.queueOverview.selectedItems = make(map[int]bool)
				m = m.rebuildQueueTable()
				return m, nil
			}
			return m, tea.Quit
		default:
			var cmd tea.Cmd
			m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) QueueOverviewView() string {
	filteredQueues := m.getFilteredQueues()

	if len(m.state.queueOverview.queues) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(styles.MediumGray)
		hintStyle := lipgloss.NewStyle().Foreground(styles.AccentColor)

		emptyMsg := lipgloss.JoinVertical(lipgloss.Center,
			emptyStyle.Render("No queues found"),
			"",
			emptyStyle.Render("Press ")+hintStyle.Render("Ctrl+N")+emptyStyle.Render(" to create a new queue"),
		)

		return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, emptyMsg)
	}

	tableView := m.state.queueOverview.table.View()

	if len(filteredQueues) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(styles.MediumGray)
		hintStyle := lipgloss.NewStyle().Foreground(styles.AccentColor)

		emptyMsg := lipgloss.JoinVertical(lipgloss.Center,
			emptyStyle.Render("No queues match your filter"),
			"",
			emptyStyle.Render("Press ")+hintStyle.Render("q")+emptyStyle.Render(" to clear the filter"),
		)

		return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, emptyMsg)
	}

	return tableView
}
