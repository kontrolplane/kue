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
    selected      int
    queue         kue.Queue
    messages      []kue.Message
    table         table.Model
    selectedRows  map[int]bool // track multi-selection by row index
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

	messages, err := kue.FetchQueueMessages(m.client, m.context, m.state.queueDetails.queue.Url, 10)
	if err != nil {
		m.error = fmt.Sprintf("Error fetching queue message(s): %v", err)
	}

	m.state.queueDetails.messages = messages
    // reset selections
    m.state.queueDetails.selectedRows = map[int]bool{}
    if m.state.queueDetails.table.Columns() == nil || len(m.state.queueDetails.table.Columns()) == 0 {
        m.state.queueDetails.table = initMessageDetailsTable()
    }
    m.refreshMessageRows()
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

// refreshMessageRows rebuilds the table rows applying selection markers.
func (m model) refreshMessageRows() {
    var rows []table.Row
    for idx, message := range m.state.queueDetails.messages {
        sel := " "
        if m.state.queueDetails.selectedRows[idx] {
            sel = "✓"
        }
        id := message.MessageID
        if len(id) > 10 {
            id = id[:10]
        }
        rows = append(rows, table.Row{
            sel+" "+id,
            message.Body,
            message.SentTimestamp,
            fmt.Sprintf("%d", len(message.Body)),
        })
    }
    if m.state.queueDetails.table.Columns() == nil || len(m.state.queueDetails.table.Columns()) == 0 {
        m.state.queueDetails.table = initMessageDetailsTable()
    }
    m.state.queueDetails.table.SetRows(rows)
    m.state.queueDetails.table.SetCursor(m.state.queueDetails.selected)
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
		case key.Matches(msg, m.keys.Select):
                // toggle selection of current row
                idx := m.state.queueDetails.selected
                if m.state.queueDetails.selectedRows == nil {
                    m.state.queueDetails.selectedRows = map[int]bool{}
                }
                if m.state.queueDetails.selectedRows[idx] {
                    delete(m.state.queueDetails.selectedRows, idx)
                } else {
                    m.state.queueDetails.selectedRows[idx] = true
                }
                m.refreshMessageRows()
            case key.Matches(msg, m.keys.Delete):
                // ctrl+d pressed, build slice of selected messages
                var msgs []kue.Message
                if len(m.state.queueDetails.selectedRows) == 0 {
                    // if none selected, default to current
                    msgs = append(msgs, m.state.queueDetails.messages[m.state.queueDetails.selected])
                } else {
                    for idx := range m.state.queueDetails.selectedRows {
                        if idx < len(m.state.queueDetails.messages) {
                            msgs = append(msgs, m.state.queueDetails.messages[idx])
                        }
                    }
                }
                m.state.queueMessageDelete.queue = m.state.queueDetails.queue
                m.state.queueMessageDelete.messages = msgs
                return m.QueueMessageDeleteSwitchPage(msg)
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

	return m.state.queueDetails.table.View()
}
