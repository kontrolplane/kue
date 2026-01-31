package tui

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/kontrolplane/kue/pkg/client"

	tea "github.com/charmbracelet/bubbletea"
	keys "github.com/kontrolplane/kue/pkg/keys"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

var mainStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func NewModel(
	projectName string,
	programName string,
) (tea.Model, error) {

	var error string
	var queues []kue.Queue
	var messages []kue.Message

	context := context.Background()

	client, err := client.CreateSqsClient(context)
	if err != nil {
		error = fmt.Sprintf("[NewModel] Couldn't create SQS client: %v", err)
	}

	queues, err = kue.ListQueuesUrls(client, context)
	if err != nil {
		error = fmt.Sprintf("[NewModel] Error listing queues: %v", err)
	}

	for i, queue := range queues {
		queue, err = kue.FetchQueueAttributes(client, context, queue.Url)
		if err != nil {
			error = fmt.Sprintf("[NewModel] Error fetching queue attributes: %v", err)
		}
		queues[i] = queue
	}

	// Initialize with default height (will be updated by first WindowSizeMsg)
	queueOverviewTable := initQueueOverviewTable(10)

	m := model{
		projectName: projectName,
		programName: programName,
		page:        queueOverview,
		context:     context,
		client:      client,

		error: error,

		keys: keys.Keys,
		help: help.New(),

		state: state{
			queueOverview: queueOverviewState{
				selected: 0,
				table:    queueOverviewTable,
				queues:   queues,
			},
			queueDetails: queueDetailsState{
				selected:        0,
				messages:        messages,
				attributesTable: "",
			},
			queueDelete: queueDeleteState{
				selected: 0,
			},
		},
	}

	var queueOverviewRows []table.Row
	for _, queue := range queues {
		queueOverviewRows = append(queueOverviewRows, table.Row{
			queue.Name,
			queue.LastModified,
			queue.ApproximateNumberOfMessages,
			queue.ApproximateNumberOfMessagesNotVisible,
			queue.ApproximateNumberOfMessagesDelayed,
		})
	}

	m.state.queueOverview.table.SetRows(queueOverviewRows)
	m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)

	log.Println("[NewModel] Model initialized")
	return m, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		log.Printf("[Update] Window size changed to %dx%d", msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height

		// Update help width
		m.help.Width = msg.Width

		// Calculate available height for content
		// Account for: header (3 lines), footer/help (variable), padding (6 lines)
		headerHeight := 3
		footerHeight := lipgloss.Height(m.help.View(m.keys))
		padding := 6
		availableHeight := msg.Height - headerHeight - footerHeight - padding

		// Ensure minimum height
		if availableHeight < 5 {
			availableHeight = 5
		}

		log.Printf("[Update] Available height for tables: %d (total: %d, header: %d, footer: %d, padding: %d)",
			availableHeight, msg.Height, headerHeight, footerHeight, padding)

		// Update queue overview table height
		cols := m.state.queueOverview.table.Columns()
		rows := m.state.queueOverview.table.Rows()
		focused := m.state.queueOverview.table.Focused()

		m.state.queueOverview.table = table.New(
			table.WithColumns(cols),
			table.WithRows(rows),
			table.WithFocused(focused),
			table.WithHeight(availableHeight),
		)
		m.state.queueOverview.table.SetStyles(defaultTableStyles())
		m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)

		// Update queue details message table height if it has been initialized
		if len(m.state.queueDetails.messagesTable.Columns()) > 0 {
			// For queue details, split available height (reserve some for attributes table)
			messageTableHeight := availableHeight - 8 // Reserve ~8 lines for attributes
			if messageTableHeight < 5 {
				messageTableHeight = 5
			}

			msgCols := m.state.queueDetails.messagesTable.Columns()
			msgRows := m.state.queueDetails.messagesTable.Rows()
			msgFocused := m.state.queueDetails.messagesTable.Focused()

			m.state.queueDetails.messagesTable = table.New(
				table.WithColumns(msgCols),
				table.WithRows(msgRows),
				table.WithFocused(msgFocused),
				table.WithHeight(messageTableHeight),
			)
			m.state.queueDetails.messagesTable.SetStyles(defaultTableStyles())
			m.state.queueDetails.messagesTable.SetCursor(m.state.queueDetails.selected)
		}

	case tea.KeyMsg:
		log.Printf("[Update] Key pressed: %s", msg.String())
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	var cmd tea.Cmd
	switch m.page {
	case queueOverview:
		m, cmd = m.QueueOverviewUpdate(msg)
	case queueDetails:
		m, cmd = m.QueueDetailsUpdate(msg)
	case queueDelete:
		m, cmd = m.QueueDeleteUpdate(msg)
	}

	return m, cmd
}

func (m model) View() string {
	log.Printf("[View] Rendering view for page: %d, queue count: %d", m.page, len(m.state.queueOverview.queues))

	var h string = formatHeader(m.projectName, m.programName, views[m.page])
	var f string = m.help.View(m.keys)
	var c string

	switch m.page {
	case queueOverview:
		c = m.QueueOverviewView()
	case queueDetails:
		c = m.QueueDetailsView()
	case queueDelete:
		c = m.QueueDeleteView()
	default:
		c = errNoPageSelected
	}

	if m.error != "" {
		log.Printf("[View] Rendering error: %s", m.error)
		c = m.ErrorView()
	}

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(h + "\n\n" + mainStyle.Render(c) + "\n\n" + f)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}
