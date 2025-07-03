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

type queueDetailsState struct {
	selected int
	queue    kue.Queue
	messages []kue.Message
	table    table.Model
}

var attributeColumnMap = map[int]string{
	0: "attribute",
	1: "value",
}

var attributeColumns []table.Column = []table.Column{
	{
		Title: attributeColumnMap[0], Width: 30,
	},
	{
		Title: attributeColumnMap[1], Width: 60,
	},
}

func initQueueDetailsTable() table.Model {
    t := table.New(
        table.WithColumns(attributeColumns),
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

func (m model) QueueDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {

	log.Println("[QueueDetailsSwitchPage]")

	messages, err := kue.FetchQueueMessages(m.client, m.context, m.state.queueDetails.queue.Url, 10)
	if err != nil {
		m.error = fmt.Sprintf("Error fetching queue message(s): %v", err)
	}

    // Build attribute rows for table
    q := m.state.queueDetails.queue
    attributeRows := []table.Row{
        {"arn", q.Arn},
        {"created", q.CreatedTimestamp},
        {"last modified", q.LastModified},
        {"delay seconds", q.DelaySeconds},
        {"max message size", q.MaxMessageSize},
        {"message retention period", q.MessageRetentionPeriod},
        {"receive msg wait", q.ReceiveMessageWaitTime},
        {"visibility timeout", q.VisibilityTimeout},
        {"approx. messages", q.ApproximateNumberOfMessages},
        {"not visible", q.ApproximateNumberOfMessagesNotVisible},
        {"delayed", q.ApproximateNumberOfMessagesDelayed},
        {"fifo queue", q.FifoQueue},
        {"content based dedup", q.ContentBasedDeduplication},
    }

    // add tags if any
    for k, v := range q.Tags {
        attributeRows = append(attributeRows, table.Row{"tag:" + k, v})
    }

    m.state.queueDetails.table.SetRows(attributeRows)
    m.state.queueDetails.table.SetCursor(0)

    // Keep messages retrieval for future use but not essential for details page yet
    m.state.queueDetails.messages = messages

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
            m.state.queueDetails.table, cmd = m.state.queueDetails.table.Update(msg)
        case key.Matches(msg, m.keys.Up):
            m.state.queueDetails.table, cmd = m.state.queueDetails.table.Update(msg)
		case key.Matches(msg, m.keys.Quit):
			return m.QueueOverviewSwitchPage(msg)
		default:
			m.state.queueDetails.table, cmd = m.state.queueDetails.table.Update(msg)
		}
	default:
		m.state.queueDetails.table, cmd = m.state.queueDetails.table.Update(msg)
	}

	return m, cmd
}

func (m model) QueueDetailsView() string {

    log.Println("[QueueDetailsView] queue:", m.state.queueDetails.queue.Name)
    return m.state.queueDetails.table.View()
}
