package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"

	tea "charm.land/bubbletea/v2"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueDetailsState holds the state for the queue details view.
type queueDetailsState struct {
	selected        int
	queue           kue.Queue
	messages        []kue.Message
	attributesTable string
	messagesTable   table.Model
	selectedItems   map[int]bool // tracks which messages are selected for bulk operations
	filter          filterModel
}

func messageColumns() []table.Column {
	const (
		numColumns     = 4
		cellPadding    = numColumns * 2 // bubbles table adds 1 char padding on each side per cell
		messageIDWidth = 34
		sentWidth      = 28
		sizeWidth      = 10
	)
	fixedWidth := messageIDWidth + sentWidth + sizeWidth
	bodyWidth := contentWidth - fixedWidth - cellPadding

	return []table.Column{
		{Title: "Message ID", Width: messageIDWidth},
		{Title: "Body", Width: bodyWidth},
		{Title: "Sent", Width: sentWidth},
		{Title: "Size", Width: sizeWidth},
	}
}

// renderAttributesPanel renders queue attributes as a two-column key-value layout
// aligned with the message table columns below it.
func renderAttributesPanel(q kue.Queue) string {
	// Derive panel widths from the message table columns so the left panel
	// aligns with "Message ID" and the right panel aligns with "Sent".
	cols := messageColumns()
	const cellPad = 2 // 1 char padding on each side per cell
	leftPanelWidth := cols[0].Width + cellPad + cols[1].Width + cellPad
	rightPanelWidth := cols[2].Width + cellPad + cols[3].Width + cellPad

	const labelWidth = 16
	leftValueWidth := leftPanelWidth - labelWidth
	rightValueWidth := rightPanelWidth - labelWidth

	labelStyle := lipgloss.NewStyle().
		Foreground(styles.MediumGray).
		Width(labelWidth)

	leftValue := lipgloss.NewStyle().
		Foreground(styles.TextLight).
		Width(leftValueWidth).
		MaxWidth(leftValueWidth)

	rightValue := lipgloss.NewStyle().
		Foreground(styles.TextLight).
		Width(rightValueWidth).
		MaxWidth(rightValueWidth)

	accentValue := lipgloss.NewStyle().
		Foreground(styles.AccentColor).
		Bold(true).
		Width(rightValueWidth).
		MaxWidth(rightValueWidth)

	leftRow := func(label, value string) string {
		return labelStyle.Render(label) + leftValue.Render(value)
	}

	rightRow := func(label, value string) string {
		return labelStyle.Render(label) + rightValue.Render(value)
	}

	rightAccentRow := func(label, value string) string {
		return labelStyle.Render(label) + accentValue.Render(value)
	}

	// Left column: identity & config — aligns with Message ID column
	leftRows := lipgloss.NewStyle().Width(leftPanelWidth).PaddingLeft(1).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			leftRow("Name", q.Name),
			leftRow("ARN", q.Arn),
			leftRow("Created", q.CreatedTimestamp),
			leftRow("Modified", q.LastModified),
			leftRow("Visibility", q.VisibilityTimeout+"s"),
		))

	// Right column: message stats — aligns with Sent column
	rightRows := lipgloss.NewStyle().Width(rightPanelWidth).PaddingLeft(1).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			rightAccentRow("Messages", q.ApproximateNumberOfMessages),
			rightRow("Not Visible", q.ApproximateNumberOfMessagesNotVisible),
			rightRow("Delayed", q.ApproximateNumberOfMessagesDelayed),
			rightRow("Delay Seconds", q.DelaySeconds),
			rightRow("Retention", formatRetention(q.MessageRetentionPeriod)),
		))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftRows, rightRows)
}

func initMessageDetailsTable(height int) table.Model {
	if height < minTableHeight {
		height = minTableHeight
	}

	t := table.New(
		table.WithColumns(messageColumns()),
		table.WithFocused(true),
		table.WithWidth(contentWidth),
		table.WithHeight(height),
	)

	t.SetStyles(styles.TableStyles())
	return t
}

func (m model) QueueDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	m = m.SwitchPage(queueDetails)
	m.loading = true
	m.loadingMsg = "Loading queue details..."
	m.state.queueDetails.selected = 0
	m.state.queueDetails.selectedItems = make(map[int]bool)
	m.state.queueDetails.filter = newFilter("Type to filter messages...")

	// Clear stale data to prevent showing old content during load
	m.state.queueDetails.attributesTable = ""
	m.state.queueDetails.messages = nil
	m.state.queueDetails.messagesTable = initMessageDetailsTable(m.getMessageTableHeight())

	return m, tea.Batch(
		commands.LoadQueueAttributes(m.context, m.client, m.state.queueDetails.queue.Url),
		commands.LoadMessages(m.context, m.client, m.state.queueDetails.queue.Url, 10),
	)
}

func (m model) getFilteredMessages() []kue.Message {
	if m.state.queueDetails.filter.text == "" {
		return m.state.queueDetails.messages
	}
	var filtered []kue.Message
	for _, msg := range m.state.queueDetails.messages {
		if m.state.queueDetails.filter.Matches(msg.MessageID) ||
			m.state.queueDetails.filter.Matches(msg.Body) {
			filtered = append(filtered, msg)
		}
	}
	return filtered
}

// QueueDetailsGoBack returns to queue details without reloading data.
func (m model) QueueDetailsGoBack(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	return m.SwitchPage(queueDetails), nil
}

func (m model) NoMessagesFound() bool {
	return m.MessagesCount() == 0
}

func (m model) MessagesCount() int {
	return len(m.state.queueDetails.messages)
}

func (m model) nextMessage() model {
	filteredMessages := m.getFilteredMessages()
	if m.state.queueDetails.selected < len(filteredMessages)-1 {
		m.state.queueDetails.selected++
	}
	return m
}

func (m model) previousMessage() model {
	if m.state.queueDetails.selected > 0 {
		m.state.queueDetails.selected--
	}
	return m
}

func (m model) toggleMessageSelection() model {
	if len(m.state.queueDetails.messages) == 0 {
		return m
	}
	idx := m.state.queueDetails.selected
	if m.state.queueDetails.selectedItems == nil {
		m.state.queueDetails.selectedItems = make(map[int]bool)
	}
	if m.state.queueDetails.selectedItems[idx] {
		delete(m.state.queueDetails.selectedItems, idx)
	} else {
		m.state.queueDetails.selectedItems[idx] = true
	}
	return m
}

func (m model) getSelectedMessages() []kue.Message {
	var messages []kue.Message
	for idx := range m.state.queueDetails.selectedItems {
		if idx < len(m.state.queueDetails.messages) {
			messages = append(messages, m.state.queueDetails.messages[idx])
		}
	}
	return messages
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	// Handle filter mode - delegate to filter model
	if m.state.queueDetails.filter.active {
		var cmd tea.Cmd
		m.state.queueDetails.filter, cmd = m.state.queueDetails.filter.Update(msg)
		m.state.queueDetails.selected = 0
		m = m.rebuildMessagesTable()
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Filter):
			var cmd tea.Cmd
			m.state.queueDetails.filter, cmd = m.state.queueDetails.filter.Activate()
			return m, cmd
		case key.Matches(msg, m.keys.Down):
			m = m.nextMessage()
			m = m.rebuildMessagesTable()
		case key.Matches(msg, m.keys.Up):
			m = m.previousMessage()
			m = m.rebuildMessagesTable()
		case key.Matches(msg, m.keys.Select):
			m = m.toggleMessageSelection()
			m = m.rebuildMessagesTable()
		case key.Matches(msg, m.keys.View):
			filteredMessages := m.getFilteredMessages()
			if len(filteredMessages) > 0 {
				selected := m.state.queueDetails.selected
				m.state.queueMessageDetails.message = filteredMessages[selected]
				m.state.queueMessageDetails.queueName = m.state.queueDetails.queue.Name
				m.state.queueMessageDetails.queueUrl = m.state.queueDetails.queue.Url
				m.state.queueMessageDetails.isFifo = m.state.queueDetails.queue.FifoQueue == "true"
				return m.QueueMessageDetailsSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Delete):
			filteredMessages := m.getFilteredMessages()
			if len(filteredMessages) > 0 {
				// If items are selected, delete selected items; otherwise delete current item
				if len(m.state.queueDetails.selectedItems) > 0 {
					m.state.queueMessageDelete.messages = m.getSelectedMessages()
				} else {
					selected := m.state.queueDetails.selected
					message := filteredMessages[selected]
					if message.ReceiptHandle != "" {
						m.state.queueMessageDelete.messages = []kue.Message{message}
					}
				}
				if len(m.state.queueMessageDelete.messages) > 0 {
					m.state.queueMessageDelete.queueUrl = m.state.queueDetails.queue.Url
					m.state.queueMessageDelete.queueName = m.state.queueDetails.queue.Name
					return m.QueueMessageDeleteSwitchPage(msg)
				}
			}
		case key.Matches(msg, m.keys.CopyToClipboard):
			if m.state.queueDetails.queue.Arn != "" {
				return m, commands.CopyToClipboard(m.state.queueDetails.queue.Arn)
			}
		case key.Matches(msg, m.keys.Purge):
			m.state.queuePurge.queue = m.state.queueDetails.queue
			m.state.queuePurge.fromOverview = false
			return m.QueuePurgeSwitchPage(msg)
		case key.Matches(msg, m.keys.Redrive):
			sourceArn := m.findSourceQueueArn(m.state.queueDetails.queue.Arn)
			if sourceArn != "" {
				m.state.queueRedrive.queue = m.state.queueDetails.queue
				m.state.queueRedrive.destinationArn = sourceArn
				m.state.queueRedrive.fromOverview = false
				return m.QueueRedriveSwitchPage(msg)
			}
			m.error = "This queue is not a dead-letter queue"
		case key.Matches(msg, m.keys.Create):
			m.state.queueMessageCreate.queueName = m.state.queueDetails.queue.Name
			m.state.queueMessageCreate.queueUrl = m.state.queueDetails.queue.Url
			m.state.queueMessageCreate.isFifo = m.state.queueDetails.queue.FifoQueue == "true"
			return m.QueueMessageCreateSwitchPage(msg)
		case key.Matches(msg, m.keys.Quit):
			// If filtering, clear filter
			if m.state.queueDetails.filter.text != "" {
				m.state.queueDetails.filter = m.state.queueDetails.filter.Clear()
				m.state.queueDetails.selected = 0
				m = m.rebuildMessagesTable()
				return m, nil
			}
			// If items are selected, clear selection instead of going back
			if len(m.state.queueDetails.selectedItems) > 0 {
				m.state.queueDetails.selectedItems = make(map[int]bool)
				m = m.rebuildMessagesTable()
				return m, nil
			}
			return m.QueueOverviewSwitchPage(msg)
		default:
			var cmd tea.Cmd
			m.state.queueDetails.messagesTable, cmd = m.state.queueDetails.messagesTable.Update(msg)
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.state.queueDetails.messagesTable, cmd = m.state.queueDetails.messagesTable.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) QueueDetailsView() string {
	// Attributes section
	var attributesView string
	if m.state.queueDetails.attributesTable != "" {
		attributesView = m.state.queueDetails.attributesTable
	} else {
		attributesView = lipgloss.NewStyle().
			Foreground(styles.MediumGray).
			Render("Loading queue attributes...")
	}

	filteredMessages := m.getFilteredMessages()
	if len(filteredMessages) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(styles.MediumGray).
			Render(fmt.Sprintf("No messages found in queue: %s", m.state.queueDetails.queue.Name))

		tableHeight := m.getMessageTableHeight()
		emptyView := lipgloss.Place(contentWidth, tableHeight, lipgloss.Center, lipgloss.Center, emptyMsg)
		return attributesView + "\n\n" + emptyView
	}

	messagesTableView := m.state.queueDetails.messagesTable.View()
	return attributesView + "\n\n" + messagesTableView
}
