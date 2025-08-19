package tui

import (
    "fmt"
    "log"
    "strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueDetailsState holds the state that is required for the queue details
// page. Besides the messages that are fetched when a user enters the page we
// also keep a pointer to the queue itself because most attributes such as the
// ARN or the message statistics are already available on that struct.
//
// We intentionally keep only a single interactive table (the messages table)
// because the attributes are rendered as a non-interactive table at the top of
// the view. This reduces the amount of key handling logic that is required and
// follows the existing UX pattern that can be seen on the queue overview page
// where the header section is purely informational.
//
// The attributesTable field is rendered but never receives focus.
// It is therefore of type string instead of table.Model.
// This keeps the implementation minimal while still providing a tidy layout.
// If we ever need the attributes table to be interactive we can replace the
// string with a dedicated table.Model.
//
// A dedicated height for the messages table can be configured here so that the
// attributes section and the help footer never overlap.
// -----------------------------------------------------------------------------
// NOTE:   the attributesTable field is generated once when entering the page
//         (see QueueDetailsSwitchPage) and kept as is while the user scrolls
//         through the list of messages.                                             
// -----------------------------------------------------------------------------

type queueDetailsState struct {
    selected        int
    queue           kue.Queue
    messages        []kue.Message
    table           table.Model // interactive messages table
    attributesTable string      // rendered, non-interactive attributes table
}

var messageColumnMap = map[int]string{
	0: "message id",
	1: "body",
	2: "sent",
	3: "size",
}

var messageColumns []table.Column = []table.Column{
	{
		Title: messageColumnMap[0], Width: 40,
	},
	{
		Title: messageColumnMap[1], Width: 60,
	},
	{
		Title: messageColumnMap[2], Width: 20,
	},
	{
		Title: messageColumnMap[3], Width: 10,
	},
}

// Helper that constructs the (non-interactive) attributes table that is rendered
// above the messages list in the queue detail view. The function returns a
// simple string that contains new-line separated rows so that the caller can
// concatenate it with the Bubbletea table view returned by the interactive
// messages table.
func renderAttributesTable(q kue.Queue) string {
    // we keep the width of the first column rather small to avoid a too wide
    // table. The values might still overflow but Bubbletea takes care of
    // clipping the content to the available terminal width.
    attrs := []struct{
        k string
        v string
    }{
        {"Name", q.Name},
        {"ARN", q.Arn},
        {"Created", q.CreatedTimestamp},
        {"Last modified", q.LastModified},
        {"Approx. msgs", q.ApproximateNumberOfMessages},
        {"In flight", q.ApproximateNumberOfMessagesNotVisible},
        {"Delayed", q.ApproximateNumberOfMessagesDelayed},
        {"Delay sec", q.DelaySeconds},
        {"Retention", q.MessageRetentionPeriod},
        {"Visibility TO", q.VisibilityTimeout},
    }

    var lines []string
    header := lipgloss.NewStyle().Bold(true).Render("attribute") + "\t" + lipgloss.NewStyle().Bold(true).Render("value")
    lines = append(lines, header)
    for _, a := range attrs {
        if a.v == "" { // skip empty values to keep table compact
            continue
        }
        lines = append(lines, fmt.Sprintf("%s\t%s", a.k, a.v))
    }

    // ensure a blank line at the end so that the messages table has some
    // spacing.
    return strings.Join(lines, "\n") + "\n\n"
}

func initMessageDetailsTable() table.Model {
	t := table.New(
		table.WithColumns(messageColumns),
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

    // refresh queue attributes first so that we always show the latest message
    // counts etc. If that fails we still allow the user to continue to the page
    // but show a warning in the footer later.
    q, err := kue.FetchQueueAttributes(m.client, m.context, m.state.queueDetails.queue.Url)
    if err != nil {
        m.error = fmt.Sprintf("Error fetching queue attributes: %v", err)
    } else {
        m.state.queueDetails.queue = q
        m.state.queueDetails.attributesTable = renderAttributesTable(q)
    }

    messages, err := kue.FetchQueueMessages(m.client, m.context, m.state.queueDetails.queue.Url, 10)
    if err != nil {
        m.error = fmt.Sprintf("Error fetching queue message(s): %v", err)
    }

    m.state.queueDetails.messages = messages

    // Build message rows for the interactive table
    var messageRows []table.Row
    for _, message := range messages {
        messageRows = append(messageRows, table.Row{
            message.MessageID,
            message.Body,
            message.SentTimestamp,
            fmt.Sprintf("%d", len(message.Body)),
        })
    }

    m.state.queueDetails.table = initMessageDetailsTable()
    m.state.queueDetails.table.SetRows(messageRows)
    m.state.queueDetails.table.SetCursor(m.state.queueDetails.selected)

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
			m.state.queueDetails.table.SetCursor(m.state.queueDetails.selected)
		case key.Matches(msg, m.keys.Up):
			m, cmd = m.previousMessage()
			m.state.queueDetails.table.SetCursor(m.state.queueDetails.selected)
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

	log.Println("[QueueDetailsView] queue:", m.state.queueDetails.queue.Name, m.state.queueDetails.messages)

	if m.NoMessagesFound() {
		return fmt.Sprintf("No messages found in queue: %s", m.state.queueDetails.queue.Name)
	}

	// var messageRows []table.Row
	// for _, message := range m.state.queueDetails.messages {
	// 	messageRows = append(messageRows, table.Row{
	// 		message.MessageID,
	// 		message.Body,
	// 		message.SentTimestamp,
	// 		fmt.Sprintf("%d", len(message.Body)),
	// 	})
	// }

	// m.state.queueDetails.table.SetRows(messageRows)
	// m.state.queueDetails.table.SetCursor(m.state.queueDetails.selected)

    // Combine the attribute information and the interactive message list. The
    // attributes table is placed on top and therefore prepended to the view.
    return m.state.queueDetails.attributesTable + m.state.queueDetails.table.View()
}
