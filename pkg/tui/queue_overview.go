package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueOverviewState holds the state for the queue listing view.
type queueOverviewState struct {
	selected      int
	queues        []kue.Queue
	loaded        bool
	table         dataTable
	selectedItems map[string]bool // queue urls selected for bulk operations
	filtering     bool
	filterInput   textinput.Model
	filterText    string
}

var queueOverviewColumns = []column{
	{title: "queue", width: 36, grow: 1, min: 16},
	{title: "type", width: 14},
	{title: "available", width: 10, right: true},
	{title: "in flight", width: 10, right: true},
	{title: "delayed", width: 8, right: true, drop: 2},
	{title: "visibility", width: 10, right: true, drop: 3},
	{title: "retention", width: 9, right: true, drop: 1},
	{title: "modified", width: 12, right: true},
}

func queueMatches(q kue.Queue, filter string) bool {
	return strings.Contains(strings.ToLower(q.Name), strings.ToLower(filter))
}

func (m model) getFilteredQueues() []kue.Queue {
	o := m.state.queueOverview
	return filterBy(o.queues, o.filterText, queueMatches)
}

// queueType renders whether a queue is standard or fifo, and marks the dead-letter queues. The
// type is padded to the width of "standard" so the dlq marks line up down the column.
func queueType(queues []kue.Queue, q kue.Queue) cell {
	kind := "standard"
	if q.FifoQueue == "true" {
		kind = "fifo"
	}
	if !isDeadLetter(queues, q) {
		return text(kind, styles.ToneBody)
	}
	return cell{styles.S(fmt.Sprintf("%-8s", kind), styles.ToneBody), styles.S(" dlq", styles.ToneWarning)}
}

func (m model) updateQueueOverviewTable() model {
	o := &m.state.queueOverview
	var rows []tableRow
	matches := filterMatches(o.filterText)
	for _, q := range m.getFilteredQueues() {
		name := cell{styles.B(q.Name, styles.ToneText)}
		if o.selectedItems[q.Url] {
			name = cell{styles.S("● ", styles.ToneAccent), styles.B(q.Name, styles.ToneText)}
		}
		name = markMatches(name, matches)
		rows = append(rows, tableRow{
			name,
			queueType(o.queues, q),
			tableCount(atoi(q.ApproximateNumberOfMessages), styles.ToneText),
			tableCount(atoi(q.ApproximateNumberOfMessagesNotVisible), styles.ToneInfo),
			tableCount(atoi(q.ApproximateNumberOfMessagesDelayed), styles.ToneBody),
			text(formatSeconds(q.VisibilityTimeout), styles.ToneMuted),
			text(formatSeconds(q.MessageRetentionPeriod), styles.ToneMuted),
			lastActivity(parseTime(q.LastModified)),
		})
	}
	o.table.empty = "no queues yet, press ctrl+n to create one"
	if len(o.queues) > 0 {
		o.table.empty = "no queues match the filter"
	}
	o.selected = setRows(&o.table, rows, o.selected)
	return m
}

func (m model) currentQueue() (kue.Queue, bool) {
	queues := m.getFilteredQueues()
	if len(queues) == 0 {
		return kue.Queue{}, false
	}
	return queues[m.state.queueOverview.selected], true
}

func (m model) toggleQueueSelection() model {
	if q, ok := m.currentQueue(); ok {
		toggle(m.state.queueOverview.selectedItems, q.Url)
	}
	return m.updateQueueOverviewTable()
}

// getSelectedQueues returns the names of the selected queues the filter shows, or the queue under
// the cursor when nothing is selected. hidden counts the selected queues the filter hides, which
// an action on the selection leaves alone.
func (m model) getSelectedQueues() (names []string, hidden int) {
	o := m.state.queueOverview
	filtered := m.getFilteredQueues()
	urls := selectedOr(o.selectedItems, filtered, func(q kue.Queue) string { return q.Url }, m.currentQueue)
	for _, q := range filtered {
		for _, url := range urls {
			if q.Url == url {
				names = append(names, q.Name)
			}
		}
	}
	if len(o.selectedItems) > 0 {
		hidden = len(o.selectedItems) - len(names)
	}
	return names, hidden
}

// clearFilterOrSelection clears the filter, or else the selection, and reports whether there
// was either to clear.
func (m model) clearFilterOrSelection() (model, bool) {
	o := &m.state.queueOverview
	switch {
	case o.filterText != "":
		o.filterText = ""
		o.filterInput.SetValue("")
		o.selected = 0
	case len(o.selectedItems) > 0:
		o.selectedItems = make(map[string]bool)
	default:
		return m, false
	}
	return m.updateQueueOverviewTable(), true
}

// QueueOverviewSwitchPage opens the queue overview and loads it.
func (m model) QueueOverviewSwitchPage() (model, tea.Cmd) {
	m.error = ""
	m = m.SwitchPage(queueOverview)
	m.loading = !m.state.queueOverview.loaded
	m.loadingMsg = "loading queues…"
	return m, commands.LoadQueues(m.context, m.client)
}

// QueueOverviewGoBack returns to the queue overview as it was left and refreshes it in the background.
func (m model) QueueOverviewGoBack() (model, tea.Cmd) {
	m.error = ""
	return m.SwitchPage(queueOverview), commands.LoadQueues(m.context, m.client)
}

func (m model) QueueOverviewUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	o := &m.state.queueOverview

	if o.filtering {
		if msg, ok := msg.(tea.KeyPressMsg); ok {
			switch {
			case key.Matches(msg, m.keys.Back):
				o.filtering = false
				o.filterText = ""
				o.filterInput.SetValue("")
				o.filterInput.Blur()
				return m.updateQueueOverviewTable(), nil
			case key.Matches(msg, m.keys.View):
				o.filtering = false
				o.filterInput.Blur()
				return m, nil
			}
		}
		o.filterInput, cmd = o.filterInput.Update(msg)
		if o.filterText != o.filterInput.Value() {
			o.filterText = o.filterInput.Value()
			o.selected = 0
		}
		return m.updateQueueOverviewTable(), cmd
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(keyMsg, m.keys.Filter):
		o.filtering = true
		return m, o.filterInput.Focus()
	case key.Matches(keyMsg, m.keys.Select):
		return m.toggleQueueSelection(), nil
	case key.Matches(keyMsg, m.keys.Refresh):
		return m, commands.LoadQueues(m.context, m.client)
	case key.Matches(keyMsg, m.keys.View):
		if q, ok := m.currentQueue(); ok {
			m.state.queueDetails.queue = q
			return m.QueueDetailsSwitchPage()
		}
	case key.Matches(keyMsg, m.keys.Create):
		return m.QueueCreateSwitchPage()
	case key.Matches(keyMsg, m.keys.Purge):
		if q, ok := m.currentQueue(); ok {
			m.state.queuePurge.queue = q
			m.state.queuePurge.fromOverview = true
			return m.QueuePurgeSwitchPage()
		}
	case key.Matches(keyMsg, m.keys.Redrive):
		if q, ok := m.currentQueue(); ok {
			return m.openRedrive(q, true)
		}
	case key.Matches(keyMsg, m.keys.Delete):
		names, hidden := m.getSelectedQueues()
		if len(names) == 0 {
			if hidden > 0 {
				return m.setStatus("the selected queues are hidden by the filter", styles.ToneWarning)
			}
			return m, nil
		}
		m.state.queueDelete.queues = names
		m.state.queueDelete.hidden = hidden
		return m.QueueDeleteSwitchPage()
	case key.Matches(keyMsg, m.keys.Back):
		m, _ = m.clearFilterOrSelection()
		return m, nil
	case key.Matches(keyMsg, m.keys.Quit):
		if m, ok := m.clearFilterOrSelection(); ok {
			return m, nil
		}
		return m, tea.Quit
	default:
		o.table = o.table.Update(keyMsg)
		o.selected = o.table.Cursor()
	}

	return m, cmd
}

func (m model) QueueOverviewView() string {
	return m.state.queueOverview.table.View()
}
