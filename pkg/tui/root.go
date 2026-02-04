package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/kontrolplane/kue/pkg/client"

	tea "github.com/charmbracelet/bubbletea"
	keys "github.com/kontrolplane/kue/pkg/keys"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/messages"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

func NewModel(
	projectName string,
	programName string,
) (tea.Model, error) {

	ctx := context.Background()

	sqsClient, err := client.CreateSqsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create SQS client: %w", err)
	}

	// Initialize with default height (will be updated by first WindowSizeMsg)
	queueOverviewTable := initQueueOverviewTable(defaultTableHeight)

	m := model{
		projectName: projectName,
		programName: programName,
		page:        queueOverview,
		context:     ctx,
		client:      sqsClient,
		loading:     true,
		loadingMsg:  "Loading queues...",

		keys: keys.Keys,
		help: help.New(),

		state: state{
			queueOverview: queueOverviewState{
				selected: 0,
				table:    queueOverviewTable,
				queues:   nil,
			},
			queueDetails: queueDetailsState{
				selected:        0,
				messages:        nil,
				attributesTable: "",
			},
			queueDelete: queueDeleteState{
				selected: 0,
			},
		},
	}

	return m, nil
}

func (m model) Init() tea.Cmd {
	// Load queues on startup
	return commands.LoadQueues(m.context, m.client)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Update all tables with consistent sizing
		m = m.resizeTables()

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}

	// Handle async message results
	case messages.QueuesLoadedMsg:
		m.loading = false
		m.loadingMsg = ""
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error loading queues: %v", msg.Err)
		} else {
			m.error = ""
			m.state.queueOverview.queues = msg.Queues
			m = m.updateQueueOverviewTable()

			// Schedule auto-refresh if on queue overview page
			if m.page == queueOverview {
				cmds = append(cmds, commands.ScheduleRefresh("queueOverview"))
			}
		}

	case messages.QueueAttributesLoadedMsg:
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error fetching queue attributes: %v", msg.Err)
		} else {
			m.state.queueDetails.queue = msg.Queue
			m.state.queueDetails.attributesTable = renderAttributesTable(msg.Queue)
		}

	case messages.MessagesLoadedMsg:
		m.loading = false
		m.loadingMsg = ""
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error fetching messages: %v", msg.Err)
		} else {
			m.state.queueDetails.messages = msg.Messages
			m = m.updateMessagesTable()

			// Schedule auto-refresh if on queue details page
			if m.page == queueDetails {
				cmds = append(cmds, commands.ScheduleRefresh("queueDetails"))
			}
		}

	case messages.QueueCreatedMsg:
		m.loading = false
		m.loadingMsg = ""
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error creating queue: %v", msg.Err)
		}
		// Refresh queue list and switch to overview
		m = m.SwitchPage(queueOverview)
		cmds = append(cmds, commands.LoadQueues(m.context, m.client))

	case messages.QueueDeletedMsg:
		m.loading = false
		m.loadingMsg = ""
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error deleting queue: %v", msg.Err)
		}
		// Reset delete selection and refresh queue list
		m.state.queueDelete.selected = 0
		m = m.SwitchPage(queueOverview)
		cmds = append(cmds, commands.LoadQueues(m.context, m.client))

	case messages.RefreshTickMsg:
		// Only refresh if still on the same page that requested it
		switch msg.Page {
		case "queueOverview":
			if m.page == queueOverview {
				cmds = append(cmds, commands.LoadQueues(m.context, m.client))
			}
		case "queueDetails":
			if m.page == queueDetails && m.state.queueDetails.queue.Url != "" {
				cmds = append(cmds, tea.Batch(
					commands.LoadQueueAttributes(m.context, m.client, m.state.queueDetails.queue.Url),
					commands.LoadMessages(m.context, m.client, m.state.queueDetails.queue.Url, 10),
				))
			}
		}
	}

	// Dispatch to page-specific Update handler
	var cmd tea.Cmd
	switch m.page {
	case queueOverview:
		m, cmd = m.QueueOverviewUpdate(msg)
	case queueDetails:
		m, cmd = m.QueueDetailsUpdate(msg)
	case queueCreate:
		m, cmd = m.QueueCreateUpdate(msg)
	case queueDelete:
		m, cmd = m.QueueDeleteUpdate(msg)
	case queueMessageDetails:
		m, cmd = m.QueueMessageDetailsUpdate(msg)
	}

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	var h string = formatHeader(m.projectName, m.programName, views[m.page])
	var f string = m.help.View(m.keys)
	var c string

	if m.loading {
		c = m.loadingMsg
		if c == "" {
			c = "Loading..."
		}
	} else {
		switch m.page {
		case queueOverview:
			c = m.QueueOverviewView()
		case queueDetails:
			c = m.QueueDetailsView()
		case queueCreate:
			c = m.QueueCreateView()
		case queueDelete:
			c = m.QueueDeleteView()
		case queueMessageDetails:
			c = m.QueueMessageDetailsView()
		default:
			c = errNoPageSelected
		}
	}

	if m.error != "" {
		c = m.ErrorView()
	}

	// Place content in a fixed-size container for consistent viewport
	fixedContent := lipgloss.Place(
		contentWidth, contentHeight,
		lipgloss.Center, lipgloss.Top,
		c,
	)

	// Wrap in border
	bordered := styles.MainBorder.Render(fixedContent)

	// Center everything in the terminal
	content := styles.ContentWrapper(m.width, m.height).
		Render(h + "\n\n" + bordered + "\n\n" + f)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}

// resizeTables updates all table heights based on fixed content dimensions.
func (m model) resizeTables() model {
	tableHeight := m.getTableHeight()

	// Update queue overview table
	cols := m.state.queueOverview.table.Columns()
	rows := m.state.queueOverview.table.Rows()
	focused := m.state.queueOverview.table.Focused()

	m.state.queueOverview.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(focused),
		table.WithHeight(tableHeight),
	)
	m.state.queueOverview.table.SetStyles(styles.TableStyles())
	m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)

	// Update queue details message table if initialized
	if len(m.state.queueDetails.messagesTable.Columns()) > 0 {
		msgCols := m.state.queueDetails.messagesTable.Columns()
		msgRows := m.state.queueDetails.messagesTable.Rows()
		msgFocused := m.state.queueDetails.messagesTable.Focused()

		m.state.queueDetails.messagesTable = table.New(
			table.WithColumns(msgCols),
			table.WithRows(msgRows),
			table.WithFocused(msgFocused),
			table.WithHeight(m.getMessageTableHeight()),
		)
		m.state.queueDetails.messagesTable.SetStyles(styles.TableStyles())
		m.state.queueDetails.messagesTable.SetCursor(m.state.queueDetails.selected)
	}

	return m
}

// updateQueueOverviewTable updates the queue overview table with current queue data.
func (m model) updateQueueOverviewTable() model {
	var rows []table.Row
	for _, queue := range m.state.queueOverview.queues {
		rows = append(rows, table.Row{
			queue.Name,
			queue.LastModified,
			queue.ApproximateNumberOfMessages,
			queue.ApproximateNumberOfMessagesNotVisible,
			queue.ApproximateNumberOfMessagesDelayed,
		})
	}

	m.state.queueOverview.table.SetRows(rows)

	// Ensure cursor is within bounds
	if m.state.queueOverview.selected >= len(m.state.queueOverview.queues) {
		m.state.queueOverview.selected = max(0, len(m.state.queueOverview.queues)-1)
	}
	m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)

	return m
}

// updateMessagesTable updates the messages table with current message data.
func (m model) updateMessagesTable() model {
	var rows []table.Row
	for _, message := range m.state.queueDetails.messages {
		rows = append(rows, table.Row{
			message.MessageID,
			message.Body,
			message.SentTimestamp,
			fmt.Sprintf("%d", len(message.Body)),
		})
	}

	m.state.queueDetails.messagesTable = initMessageDetailsTable(m.getMessageTableHeight())
	m.state.queueDetails.messagesTable.SetRows(rows)

	// Ensure cursor is within bounds
	if m.state.queueDetails.selected >= len(m.state.queueDetails.messages) {
		m.state.queueDetails.selected = max(0, len(m.state.queueDetails.messages)-1)
	}
	m.state.queueDetails.messagesTable.SetCursor(m.state.queueDetails.selected)

	return m
}
