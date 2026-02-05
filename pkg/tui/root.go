package tui

import (
	"context"
	"fmt"

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

	sqsClient, awsInfo, err := client.CreateSqsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("couldn't create SQS client: %w", err)
	}

	queueOverviewTable := initQueueOverviewTable(defaultTableHeight)

	m := model{
		projectName: projectName,
		programName: programName,
		page:        queueOverview,
		context:     ctx,
		client:      sqsClient,
		awsInfo:     awsInfo,
		loading:     true,
		loadingMsg:  "Loading queues...",

		keys: keys.Keys,

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
	return commands.LoadQueues(m.context, m.client)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m = m.resizeTables()

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

	case messages.QueuesLoadedMsg:
		m.loading = false
		m.loadingMsg = ""
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error loading queues: %v", msg.Err)
		} else {
			m.error = ""
			m.state.queueOverview.queues = msg.Queues
			m = m.updateQueueOverviewTable()
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
		m = m.SwitchPage(queueOverview)
		cmds = append(cmds, commands.LoadQueues(m.context, m.client))

	case messages.QueueDeletedMsg:
		m.loading = false
		m.loadingMsg = ""
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error deleting queue: %v", msg.Err)
		}
		m.state.queueDelete.selected = 0
		m = m.SwitchPage(queueOverview)
		cmds = append(cmds, commands.LoadQueues(m.context, m.client))

	case messages.MessageDeletedMsg:
		m.loading = false
		m.loadingMsg = ""
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error deleting message: %v", msg.Err)
		} else {
			queueUrl := m.state.queueDetails.queue.Url
			if m.page != queueDetails {
				m = m.SwitchPage(queueDetails)
			}
			if m.state.queueDetails.selected >= len(m.state.queueDetails.messages)-1 && m.state.queueDetails.selected > 0 {
				m.state.queueDetails.selected--
			}
			cmds = append(cmds, tea.Batch(
				commands.LoadQueueAttributes(m.context, m.client, queueUrl),
				commands.LoadMessages(m.context, m.client, queueUrl, 10),
			))
		}

	case messages.MessageCreatedMsg:
		m.loading = false
		m.loadingMsg = ""
		if msg.Err != nil {
			m.error = fmt.Sprintf("Error sending message: %v", msg.Err)
		} else {
			queueUrl := m.state.queueMessageCreate.queueUrl
			m = m.SwitchPage(queueDetails)
			cmds = append(cmds, tea.Batch(
				commands.LoadQueueAttributes(m.context, m.client, queueUrl),
				commands.LoadMessages(m.context, m.client, queueUrl, 10),
			))
		}

	case messages.RefreshTickMsg:
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
	case queueMessageDelete:
		m, cmd = m.QueueMessageDeleteUpdate(msg)
	case queueMessageCreate:
		m, cmd = m.QueueMessageCreateUpdate(msg)
	}

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	h := formatHeader(m.projectName, m.programName, views[m.page], m.awsInfo)
	f := m.renderShortHelp()
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
		case queueMessageDelete:
			c = m.QueueMessageDeleteView()
		case queueMessageCreate:
			c = m.QueueMessageCreateView()
		default:
			c = errNoPageSelected
		}
	}

	if m.error != "" {
		c = m.ErrorView()
	}

	fixedContent := lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Top, c)
	bordered := styles.MainBorder.Render(fixedContent)
	mainView := h + "\n\n" + bordered + "\n\n" + f

	if m.showHelp {
		mainView = m.renderHelpOverlay(mainView)
	}

	return styles.ContentWrapper(m.width, m.height).Render(mainView)
}

func (m model) renderShortHelp() string {
	helpStyle := lipgloss.NewStyle().Foreground(styles.MediumGray)
	return helpStyle.Render("? help • q quit")
}

func (m model) renderHelpOverlay(background string) string {
	helpContent := m.renderHelpContent()

	overlay := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.AccentColor).
		Padding(1, 3).
		Background(lipgloss.Color("235")).
		Render(helpContent)

	return lipgloss.Place(contentWidth+4, contentHeight+10,
		lipgloss.Center, lipgloss.Center,
		overlay,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("235")),
	)
}

func (m model) renderHelpContent() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(styles.AccentColor).
		Bold(true).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(styles.TextLight).
		Width(15)

	descStyle := lipgloss.NewStyle().
		Foreground(styles.MediumGray)

	row := func(key, desc string) string {
		return keyStyle.Render(key) + descStyle.Render(desc)
	}

	navigation := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Navigation"),
		row("↑/k", "move up"),
		row("↓/j", "move down"),
		row("←/h", "move left"),
		row("→/l", "move right"),
		row("enter", "view/select"),
	)

	actions := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Actions"),
		row("ctrl+n", "create new"),
		row("ctrl+d", "delete"),
		row("/", "filter"),
		row("q/esc", "back/quit"),
		row("?", "toggle help"),
	)

	columns := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().MarginRight(4).Render(navigation),
		actions,
	)

	footer := lipgloss.NewStyle().
		Foreground(styles.DarkGray).
		MarginTop(1).
		Render("Press any key to close")

	return lipgloss.JoinVertical(lipgloss.Center, columns, footer)
}

func (m model) resizeTables() model {
	tableHeight := m.getTableHeight()
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

func (m model) updateQueueOverviewTable() model {
	var rows []table.Row
	for _, queue := range m.state.queueOverview.queues {
		queueType := "standard"
		if queue.FifoQueue == "true" {
			queueType = "fifo"
		}
		visibility := queue.VisibilityTimeout + "s"
		retention := formatRetention(queue.MessageRetentionPeriod)

		rows = append(rows, table.Row{
			queue.Name,
			queueType,
			centerText(queue.ApproximateNumberOfMessages, 10),
			centerText(queue.ApproximateNumberOfMessagesNotVisible, 10),
			centerText(queue.ApproximateNumberOfMessagesDelayed, 10),
			centerText(visibility, 10),
			centerText(retention, 10),
			queue.LastModified,
		})
	}

	m.state.queueOverview.table.SetRows(rows)
	if m.state.queueOverview.selected >= len(m.state.queueOverview.queues) {
		m.state.queueOverview.selected = max(0, len(m.state.queueOverview.queues)-1)
	}
	m.state.queueOverview.table.SetCursor(m.state.queueOverview.selected)

	return m
}

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
	if m.state.queueDetails.selected >= len(m.state.queueDetails.messages) {
		m.state.queueDetails.selected = max(0, len(m.state.queueDetails.messages)-1)
	}
	m.state.queueDetails.messagesTable.SetCursor(m.state.queueDetails.selected)

	return m
}

func formatRetention(seconds string) string {
	if seconds == "" {
		return "-"
	}

	var secs int
	if _, err := fmt.Sscanf(seconds, "%d", &secs); err != nil {
		return seconds
	}

	days := secs / 86400
	if days > 0 {
		return fmt.Sprintf("%dd", days)
	}

	hours := secs / 3600
	if hours > 0 {
		return fmt.Sprintf("%dh", hours)
	}

	return fmt.Sprintf("%ds", secs)
}

func centerText(text string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(text)
}
