package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
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
}

// Message table column definitions.
var messageColumnMap = map[int]string{
	0: "message identifier",
	1: "body",
	2: "sent timestamp",
	3: "size",
}

var messageColumns []table.Column = []table.Column{
	{
		Title: messageColumnMap[0], Width: 30,
	},
	{
		Title: messageColumnMap[1], Width: 60,
	},
	{
		Title: messageColumnMap[2], Width: 30,
	},
	{
		Title: messageColumnMap[3], Width: 10,
	},
}

func renderAttributesTable(q kue.Queue) string {
	columnsLeft := []table.Column{
		{Title: "attribute", Width: 30},
		{Title: "value", Width: 60},
	}

	columnsRight := []table.Column{
		{Title: "attribute", Width: 30},
		{Title: "value", Width: 10},
	}

	rowsLeft := []table.Row{
		{"name", q.Name},
		{"arn", q.Arn},
		{"created at", q.CreatedTimestamp},
		{"last modified", q.LastModified},
		{"visibility timeout", q.VisibilityTimeout},
	}

	rowsRight := []table.Row{
		{"number of messages", q.ApproximateNumberOfMessages},
		{"number not visible", q.ApproximateNumberOfMessagesNotVisible},
		{"number delayed", q.ApproximateNumberOfMessagesDelayed},
		{"delay seconds", q.DelaySeconds},
		{"retention period", q.MessageRetentionPeriod},
	}

	leftTable := table.New(
		table.WithColumns(columnsLeft),
		table.WithRows(rowsLeft),
		table.WithFocused(false),
		table.WithHeight(len(rowsLeft)+1),
	)

	rightTable := table.New(
		table.WithColumns(columnsRight),
		table.WithRows(rowsRight),
		table.WithFocused(false),
		table.WithHeight(len(rowsRight)+1),
	)

	leftTable.SetStyles(styles.AttributesTableStyles())
	rightTable.SetStyles(styles.AttributesTableStyles())

	leftView := lipgloss.NewStyle().Render(stripViewBeforeToken(leftTable.View(), rowsLeft[0][0]))
	rightView := lipgloss.NewStyle().Render(stripViewBeforeToken(rightTable.View(), rowsRight[0][0]))

	return lipgloss.JoinHorizontal(lipgloss.Top, leftView, rightView)
}

// stripViewBeforeToken removes everything before the first line that contains
// the provided token. This effectively drops the header and any blank line
// above the first data row.
func stripViewBeforeToken(view string, token string) string {
	lines := strings.Split(view, "\n")
	start := 0
	for i, line := range lines {
		if strings.Contains(line, token) {
			start = i
			break
		}
	}
	return strings.Join(lines[start:], "\n")
}

func renderMessagesTableHeader() string {
	headerStyle := styles.TableStyles().Header

	var headers []string
	for _, col := range messageColumns {
		headers = append(headers, headerStyle.Width(col.Width).Render(col.Title))
	}
	header := lipgloss.JoinHorizontal(lipgloss.Top, headers...)

	// Create divider
	totalWidth := 0
	for _, col := range messageColumns {
		totalWidth += col.Width
	}
	divider := headerStyle.Render(strings.Repeat("─", totalWidth))

	return header + "\n" + divider
}

func initMessageDetailsTable(height int) table.Model {
	if height < minTableHeight {
		height = minTableHeight
	}

	t := table.New(
		table.WithColumns(messageColumns),
		table.WithFocused(true),
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

	// Clear stale data to prevent showing old content during load
	m.state.queueDetails.attributesTable = ""
	m.state.queueDetails.messages = nil
	m.state.queueDetails.messagesTable = initMessageDetailsTable(m.getMessageTableHeight())

	return m, tea.Batch(
		commands.LoadQueueAttributes(m.context, m.client, m.state.queueDetails.queue.Url),
		commands.LoadMessages(m.context, m.client, m.state.queueDetails.queue.Url, 10),
	)
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

func (m model) nextMessage() (model, tea.Cmd) {
	if m.state.queueDetails.selected < len(m.state.queueDetails.messages)-1 {
		m.state.queueDetails.selected++
	}
	return m, nil
}

func (m model) previousMessage() (model, tea.Cmd) {
	if m.state.queueDetails.selected > 0 {
		m.state.queueDetails.selected--
	}
	return m, nil
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down):
			m, cmd = m.nextMessage()
			m.state.queueDetails.messagesTable.SetCursor(m.state.queueDetails.selected)
		case key.Matches(msg, m.keys.Up):
			m, cmd = m.previousMessage()
			m.state.queueDetails.messagesTable.SetCursor(m.state.queueDetails.selected)
		case key.Matches(msg, m.keys.View):
			if len(m.state.queueDetails.messages) > 0 {
				selected := m.state.queueDetails.selected
				m.state.queueMessageDetails.message = m.state.queueDetails.messages[selected]
				m.state.queueMessageDetails.queueName = m.state.queueDetails.queue.Name
				m.state.queueMessageDetails.queueUrl = m.state.queueDetails.queue.Url
				return m.QueueMessageDetailsSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.DeleteMessage):
			if len(m.state.queueDetails.messages) > 0 {
				selected := m.state.queueDetails.selected
				message := m.state.queueDetails.messages[selected]
				if message.ReceiptHandle != "" {
					m.state.queueMessageDelete.message = message
					m.state.queueMessageDelete.queueUrl = m.state.queueDetails.queue.Url
					m.state.queueMessageDelete.queueName = m.state.queueDetails.queue.Name
					return m.QueueMessageDeleteSwitchPage(msg)
				}
			}
		case key.Matches(msg, m.keys.Create):
			m.state.queueMessageCreate.queueName = m.state.queueDetails.queue.Name
			m.state.queueMessageCreate.queueUrl = m.state.queueDetails.queue.Url
			m.state.queueMessageCreate.isFifo = m.state.queueDetails.queue.FifoQueue == "true"
			return m.QueueMessageCreateSwitchPage(msg)
		case key.Matches(msg, m.keys.Quit):
			return m.QueueOverviewSwitchPage(msg)
		default:
			m.state.queueDetails.messagesTable, cmd = m.state.queueDetails.messagesTable.Update(msg)
		}
	default:
		m.state.queueDetails.messagesTable, cmd = m.state.queueDetails.messagesTable.Update(msg)
	}

	return m, cmd
}

func (m model) QueueDetailsView() string {
	var attributesTableView string
	if m.state.queueDetails.attributesTable != "" {
		attributesTableView = m.state.queueDetails.attributesTable
	} else {
		attributesTableView = lipgloss.NewStyle().
			Foreground(styles.MediumGray).
			Render("Loading queue attributes...")
	}

	if m.NoMessagesFound() {
		header := renderMessagesTableHeader()

		emptyMsg := lipgloss.NewStyle().
			Foreground(styles.MediumGray).
			Render(fmt.Sprintf("No messages found in queue: %s", m.state.queueDetails.queue.Name))

		// Center the message in the remaining table area
		tableHeight := m.getMessageTableHeight() - 2
		centeredMsg := lipgloss.Place(contentWidth, tableHeight,
			lipgloss.Center, lipgloss.Center,
			emptyMsg)

		return attributesTableView + "\n\n" + header + "\n" + centeredMsg
	}

	messagesTableView := m.state.queueDetails.messagesTable.View()
	return attributesTableView + "\n\n" + messagesTableView
}
