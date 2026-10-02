package tui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueDetailsState holds the state for the queue details view.
type queueDetailsState struct {
	queue            kue.Queue
	attributesLoaded bool
	messages         []kue.Message
	messagesLoaded   bool
	messagesTable    dataTable
	selected         int
	selectedItems    map[string]bool // message ids selected for bulk operations
	filtering        bool
	filterInput      textinput.Model
	filterText       string
}

var messageColumns = []column{
	{title: "message id", width: 36, drop: 1},
	{title: "body", width: 56, grow: 1},
	{title: "sent", width: 12, right: true},
	{title: "receives", width: 8, right: true},
	{title: "size", width: 9, right: true},
}

const attributeLabelWidth = 16

type attribute struct {
	label string
	value cell
}

func attributeColumn(width int, rows []attribute) string {
	label := styles.Fg(styles.ToneFaint).Width(attributeLabelWidth)
	lines := make([]string, len(rows))
	for i, r := range rows {
		lines[i] = label.Render(r.label) + renderCell(r.value, column{width: width - attributeLabelWidth}, nil)
	}
	return lipgloss.NewStyle().Width(width).Render(strings.Join(lines, "\n"))
}

// queueName reads the queue name from its arn, arn:aws:sqs:region:account:name.
func queueName(arn string) string {
	return arn[strings.LastIndex(arn, ":")+1:]
}

// deadLetter describes where a queue sends the messages it fails to process, or which queues
// send theirs to it.
func deadLetter(queues []kue.Queue, q kue.Queue) cell {
	if q.DeadLetterTargetARN != "" {
		c := cell{styles.S("→ ", styles.ToneFaint), styles.S(queueName(q.DeadLetterTargetARN), styles.ToneBody)}
		var policy map[string]any
		if json.Unmarshal([]byte(q.RedrivePolicy), &policy) == nil {
			if n, ok := policy["maxReceiveCount"]; ok {
				c = append(c, sep(), styles.S(fmt.Sprintf("after %v receives", n), styles.ToneMuted))
			}
		}
		return c
	}
	var sources []string
	for _, o := range queues {
		if q.Arn != "" && o.DeadLetterTargetARN == q.Arn {
			sources = append(sources, o.Name)
		}
	}
	if len(sources) > 0 {
		return cell{styles.S("dead-letter queue of ", styles.ToneWarning), styles.S(strings.Join(sources, ", "), styles.ToneBody)}
	}
	return text("-", styles.ToneFaint)
}

func renderAttributesTable(queues []kue.Queue, q kue.Queue) string {
	kind := cell{styles.S("standard", styles.ToneBody)}
	if q.FifoQueue == "true" {
		kind = cell{styles.S("fifo", styles.ToneBody)}
		if q.ContentBasedDeduplication == "true" {
			kind = append(kind, sep(), styles.S("content based deduplication", styles.ToneBody))
		}
	}

	tags := text("-", styles.ToneFaint)
	if len(q.Tags) > 0 {
		names := make([]string, 0, len(q.Tags))
		for k, v := range q.Tags {
			names = append(names, k+"="+v)
		}
		slices.Sort(names)
		tags = text(strings.Join(names, ", "), styles.ToneBody)
	}

	created, modified := parseTime(q.CreatedTimestamp), parseTime(q.LastModified)
	timestamp := func(t time.Time) cell {
		if t.IsZero() {
			return text("-", styles.ToneFaint)
		}
		return cell{styles.S(formatTime(t), styles.ToneBody), styles.S("  "+formatAgo(t), styles.ToneFaint)}
	}

	leftWidth := (contentWidth - 6) * 53 / 100
	left := attributeColumn(leftWidth, []attribute{
		{"arn", text(orDash(q.Arn), styles.ToneBody)},
		{"type", kind},
		{"dead-letter", deadLetter(queues, q)},
		{"created", timestamp(created)},
		{"modified", timestamp(modified)},
		{"tags", tags},
	})

	seconds := func(s string) cell { return text(formatSeconds(s), styles.ToneBody) }
	maxSize := text("-", styles.ToneFaint)
	if n, err := strconv.ParseUint(q.MaxMessageSize, 10, 64); err == nil {
		maxSize = text(formatBytes(n), styles.ToneBody)
	}
	right := attributeColumn(contentWidth-6-leftWidth, []attribute{
		{"messages", cell{
			styles.B(formatCount(atoi(q.ApproximateNumberOfMessages)), styles.ToneText), styles.S(" available", styles.ToneFaint), sep(),
			styles.S(formatCount(atoi(q.ApproximateNumberOfMessagesNotVisible)), styles.ToneInfo), styles.S(" in flight", styles.ToneFaint), sep(),
			styles.S(formatCount(atoi(q.ApproximateNumberOfMessagesDelayed)), styles.ToneBody), styles.S(" delayed", styles.ToneFaint),
		}},
		{"visibility", seconds(q.VisibilityTimeout)},
		{"retention", seconds(q.MessageRetentionPeriod)},
		{"delivery delay", seconds(q.DelaySeconds)},
		{"receive wait", seconds(q.ReceiveMessageWaitTime)},
		{"max size", maxSize},
	})

	return lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
}

func (m model) QueueDetailsSwitchPage() (model, tea.Cmd) {
	m.error = ""
	m = m.SwitchPage(queueDetails)
	m.loading = true
	m.loadingMsg = "loading queue details…"

	d := &m.state.queueDetails
	d.attributesLoaded = false
	d.messages = nil
	d.messagesLoaded = false
	d.selected = 0
	d.selectedItems = make(map[string]bool)
	d.filtering = false
	d.filterText = ""
	d.filterInput.SetValue("")
	d.messagesTable = newDataTable(messageColumns, contentWidth-4, m.getMessageTableHeight())
	m = m.updateMessagesTable()

	return m, commands.LoadQueueDetails(m.context, m.client, d.queue.Url)
}

// queueDetailsGoBack returns to the queue details without receiving the messages again, which
// would count towards their receive count.
func (m model) queueDetailsGoBack() (model, tea.Cmd) {
	m.error = ""
	m = m.SwitchPage(queueDetails)
	return m.scheduleRefresh()
}

// QueueDetailsReload returns to the queue details and reloads them in the background, after
// something changed its messages.
func (m model) QueueDetailsReload() (model, tea.Cmd) {
	m.error = ""
	m = m.SwitchPage(queueDetails)
	return m, commands.LoadQueueDetails(m.context, m.client, m.state.queueDetails.queue.Url)
}

func messageMatches(msg kue.Message, filter string) bool {
	filter = strings.ToLower(filter)
	return strings.Contains(strings.ToLower(msg.MessageID), filter) ||
		strings.Contains(strings.ToLower(msg.Body), filter)
}

func (m model) getFilteredMessages() []kue.Message {
	d := m.state.queueDetails
	return filterBy(d.messages, d.filterText, messageMatches)
}

// bodyPreview renders a body on one line, tinted, with the text the filter matches set apart.
func bodyPreview(body string, matches *regexp.Regexp) cell {
	if body == "" {
		return text("empty", styles.ToneFaint)
	}
	return markMatches(tint(preview(body, 200)), matches)
}

func (m model) updateMessagesTable() model {
	d := &m.state.queueDetails
	var rows []tableRow
	matches := filterMatches(d.filterText)
	for _, msg := range m.getFilteredMessages() {
		id := cell{styles.S(msg.MessageID, styles.ToneMuted)}
		if d.selectedItems[msg.MessageID] {
			id = cell{styles.S("● ", styles.ToneAccent), styles.S(msg.MessageID, styles.ToneText)}
		}
		id = markMatches(id, matches)
		receives := atoi(msg.ReceiveCount)
		receivesTone := styles.ToneBody
		if receives > 1 {
			receivesTone = styles.ToneWarning
		}
		rows = append(rows, tableRow{
			id,
			bodyPreview(msg.Body, matches),
			lastActivity(parseTime(msg.SentTimestamp)),
			tableCount(receives, receivesTone),
			text(formatBytes(uint64(len(msg.Body))), styles.ToneMuted),
		})
	}
	switch {
	case !d.messagesLoaded:
		d.messagesTable.empty = "loading messages…"
	case len(d.messages) == 0:
		d.messagesTable.empty = fmt.Sprintf("no messages available in %s", d.queue.Name)
	default:
		d.messagesTable.empty = "no messages match the filter"
	}
	d.selected = setRows(&d.messagesTable, rows, d.selected)
	return m
}

func (m model) currentMessage() (kue.Message, bool) {
	msgs := m.getFilteredMessages()
	if len(msgs) == 0 {
		return kue.Message{}, false
	}
	return msgs[m.state.queueDetails.selected], true
}

// getSelectedMessages returns the selected messages, or the message under the cursor when nothing is selected.
func (m model) getSelectedMessages() []kue.Message {
	d := m.state.queueDetails
	ids := selectedOr(d.selectedItems, d.messages, func(msg kue.Message) string { return msg.MessageID }, m.currentMessage)
	var msgs []kue.Message
	for _, msg := range d.messages {
		if slices.Contains(ids, msg.MessageID) {
			msgs = append(msgs, msg)
		}
	}
	return msgs
}

// clearDetailsFilterOrSelection clears the filter, or else the selection, and reports whether
// there was either to clear.
func (m model) clearDetailsFilterOrSelection() (model, bool) {
	d := &m.state.queueDetails
	switch {
	case d.filterText != "":
		d.filterText = ""
		d.filterInput.SetValue("")
		d.selected = 0
	case len(d.selectedItems) > 0:
		d.selectedItems = make(map[string]bool)
	default:
		return m, false
	}
	return m.updateMessagesTable(), true
}

// openMessageCreate opens the form to send a message to the queue of the details view.
func (m model) openMessageCreate() (model, tea.Cmd) {
	q := m.state.queueDetails.queue
	c := &m.state.queueMessageCreate
	c.queueName = q.Name
	c.queueUrl = q.Url
	c.isFifo = q.FifoQueue == "true"
	return m.QueueMessageCreateSwitchPage()
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	d := &m.state.queueDetails

	if d.filtering {
		if msg, ok := msg.(tea.KeyPressMsg); ok {
			switch {
			case key.Matches(msg, m.keys.Back):
				d.filtering = false
				d.filterText = ""
				d.filterInput.SetValue("")
				d.filterInput.Blur()
				return m.updateMessagesTable(), nil
			case key.Matches(msg, m.keys.View):
				d.filtering = false
				d.filterInput.Blur()
				return m, nil
			}
		}
		d.filterInput, cmd = d.filterInput.Update(msg)
		if d.filterText != d.filterInput.Value() {
			d.filterText = d.filterInput.Value()
			d.selected = 0
		}
		return m.updateMessagesTable(), cmd
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(keyMsg, m.keys.Filter):
		d.filtering = true
		return m, d.filterInput.Focus()
	case key.Matches(keyMsg, m.keys.Select):
		if msg, ok := m.currentMessage(); ok {
			toggle(d.selectedItems, msg.MessageID)
		}
		return m.updateMessagesTable(), nil
	case key.Matches(keyMsg, m.keys.Refresh):
		return m, commands.LoadQueueDetails(m.context, m.client, d.queue.Url)
	case key.Matches(keyMsg, m.keys.View):
		if message, ok := m.currentMessage(); ok {
			md := &m.state.queueMessageDetails
			md.message = message
			md.queueName = d.queue.Name
			md.isFifo = d.queue.FifoQueue == "true"
			return m.QueueMessageDetailsSwitchPage()
		}
	case key.Matches(keyMsg, m.keys.Delete):
		if msgs := m.getSelectedMessages(); len(msgs) > 0 {
			md := &m.state.queueMessageDelete
			md.messages = msgs
			md.queueUrl = d.queue.Url
			md.queueName = d.queue.Name
			md.fromDetails = false
			return m.QueueMessageDeleteSwitchPage()
		}
	case key.Matches(keyMsg, m.keys.CopyToClipboard):
		if d.queue.Arn != "" {
			return m, commands.CopyToClipboard(d.queue.Arn)
		}
	case key.Matches(keyMsg, m.keys.Purge):
		m.state.queuePurge.queue = d.queue
		m.state.queuePurge.fromOverview = false
		return m.QueuePurgeSwitchPage()
	case key.Matches(keyMsg, m.keys.Redrive):
		return m.openRedrive(d.queue, false)
	case key.Matches(keyMsg, m.keys.Create):
		return m.openMessageCreate()
	case key.Matches(keyMsg, m.keys.Quit, m.keys.Back):
		if m, ok := m.clearDetailsFilterOrSelection(); ok {
			return m, nil
		}
		return m.QueueOverviewGoBack()
	default:
		d.messagesTable = d.messagesTable.Update(keyMsg)
		d.selected = d.messagesTable.Cursor()
	}

	return m, cmd
}

func (m model) QueueDetailsView() string {
	d := m.state.queueDetails

	attributes := styles.Muted("loading queue attributes…")
	if d.attributesLoaded {
		attributes = renderAttributesTable(m.state.queueOverview.queues, d.queue)
	}

	return lipgloss.PlaceHorizontal(contentWidth-4, lipgloss.Left, attributes) + "\n\n" +
		lipgloss.PlaceHorizontal(contentWidth-4, lipgloss.Left, d.messagesTable.View())
}
