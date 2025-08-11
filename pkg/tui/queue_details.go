package tui

import (
    "log"

    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/table"
    "github.com/charmbracelet/lipgloss"

    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueDetailsState keeps state for the queue details page
// We only need to keep the selected row, the queue information
// and the attribute table used for rendering.
type queueDetailsState struct {
    selected int
    queue    kue.Queue
    table    table.Model
}

// attributeColumns defines two columns: the attribute name and its value.
var attributeColumns = []table.Column{
    {Title: "attribute", Width: 35},
    {Title: "value", Width: 80},
}

// initQueueAttributeTable initialises a table with default styles that we
// reuse every time we enter the details page.
func initQueueAttributeTable() table.Model {
    t := table.New(
        table.WithColumns(attributeColumns),
        table.WithFocused(true),
        table.WithHeight(15),
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

// QueueDetailsSwitchPage is called when switching from the queue overview to
// the details page. We (re)fetch the attributes for the selected queue so that
// the information shown is always up to date.
func (m model) QueueDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {

    log.Println("[QueueDetailsSwitchPage]")

    // (re)fetch attributes
    queue, err := kue.FetchQueueAttributes(m.client, m.context, m.state.queueDetails.queue.Url)
    if err != nil {
        m.error = err.Error()
    } else {
        m.state.queueDetails.queue = queue
    }

    // lazily initialise the table the first time we visit the page
    if len(m.state.queueDetails.table.Columns()) == 0 {
        m.state.queueDetails.table = initQueueAttributeTable()
    }

    // Build rows from the queue struct fields that are useful to the user
    q := m.state.queueDetails.queue
    rows := []table.Row{
        {"name", q.Name},
        {"arn", q.Arn},
        {"created", q.CreatedTimestamp},
        {"last modified", q.LastModified},
        {"delay seconds", q.DelaySeconds},
        {"max message size", q.MaxMessageSize},
        {"message retention", q.MessageRetentionPeriod},
        {"receive wait time", q.ReceiveMessageWaitTime},
        {"visibility timeout", q.VisibilityTimeout},
        {"available messages", q.ApproximateNumberOfMessages},
        {"in flight messages", q.ApproximateNumberOfMessagesNotVisible},
        {"delayed messages", q.ApproximateNumberOfMessagesDelayed},
        {"fifo queue", q.FifoQueue},
        {"content-based deduplication", q.ContentBasedDeduplication},
    }

    m.state.queueDetails.table.SetRows(rows)
    m.state.queueDetails.table.SetCursor(m.state.queueDetails.selected)

    return m.SwitchPage(queueDetails), nil
}

// QueueDetailsUpdate handles key events inside the details page. We only need
// to support navigating up/down the table and quitting back to the overview.
func (m model) QueueDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Down):
            if m.state.queueDetails.selected < len(m.state.queueDetails.table.Rows())-1 {
                m.state.queueDetails.selected++
            }
            m.state.queueDetails.table.SetCursor(m.state.queueDetails.selected)
        case key.Matches(msg, m.keys.Up):
            if m.state.queueDetails.selected > 0 {
                m.state.queueDetails.selected--
            }
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

// QueueDetailsView renders the attribute table.
func (m model) QueueDetailsView() string {
    log.Println("[QueueDetailsView] queue:", m.state.queueDetails.queue.Name)
    return m.state.queueDetails.table.View()
}
