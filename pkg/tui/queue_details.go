package tui

import (
    "encoding/json"
    "fmt"
    "log"
    "strconv"

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

var messageColumnMap = map[int]string{
    0: "message id",
    1: "body",
    2: "sent",
    3: "size",
    4: "dlq",
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
    {
        Title: messageColumnMap[4], Width: 10,
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
    // (re)initialize the table every time we enter the page so that
    // refreshed column definitions are applied.
    m.state.queueDetails.table = initMessageDetailsTable()

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

// parseMaxReceiveCount extracts the maxReceiveCount from a RedrivePolicy string (JSON)
func parseMaxReceiveCount(policy string) int {
    var max int
    // policy JSON like {"deadLetterTargetArn":"arn:...","maxReceiveCount":"5"}
    type rp struct {
        MaxReceiveCount string `json:"maxReceiveCount"`
    }
    var tmp rp
    if err := json.Unmarshal([]byte(policy), &tmp); err == nil {
        max, _ = strconv.Atoi(tmp.MaxReceiveCount)
    }
    return max
}

func atoi(s string) int {
    i, _ := strconv.Atoi(s)
    return i
}

func (m model) QueueDetailsView() string {

	log.Println("[QueueDetailsView] queue:", m.state.queueDetails.queue.Name, m.state.queueDetails.messages)

	if m.NoMessagesFound() {
		return fmt.Sprintf("No messages found in queue: %s", m.state.queueDetails.queue.Name)
	}

    var messageRows []table.Row
    for _, message := range m.state.queueDetails.messages {
        dlqWarn := ""
        if m.state.queueDetails.queue.RedrivePolicy != "" {
            maxReceive := parseMaxReceiveCount(m.state.queueDetails.queue.RedrivePolicy)
            if maxReceive > 0 {
                rc := atoi(message.ReceiveCount)
                if rc >= maxReceive-1 {
                    dlqWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(fmt.Sprintf("%d/%d", rc, maxReceive))
                }
            }
        }
        messageRows = append(messageRows, table.Row{
            message.MessageID,
            message.Body,
            message.SentTimestamp,
            fmt.Sprintf("%d", len(message.Body)),
            dlqWarn,
        })
    }

    m.state.queueDetails.table.SetRows(messageRows)
    m.state.queueDetails.table.SetCursor(m.state.queueDetails.selected)

    return m.state.queueDetails.table.View()
}
