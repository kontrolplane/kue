package tui

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/kontrolplane/kue/pkg/client"
	"github.com/kontrolplane/kue/pkg/tui/commands"

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

	queueOverviewTable := initQueueOverviewTable()

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
				selected: 0,
				messages: messages,
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

	case tea.KeyMsg:
		log.Printf("[Update] Key pressed: %s", msg.String())
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	// Handle global messages (spinner/toast)
	switch t := msg.(type) {
	case QueueCreateResultMsg:
		m.creatingQueue = false
		if t.Err != nil {
			m.toast = fmt.Sprintf("❌ %v", t.Err)
		} else {
			m.toast = "✅ Queue created!"
			// Prepend new queue to slice or refresh list
			m.state.queueOverview.queues = append([]kue.Queue{t.Queue}, m.state.queueOverview.queues...)
		}
		return m, commands.ClearToastCmd(3 * time.Second)
	case ToastClearMsg:
		m.toast = ""
	}

	// update spinner if active
	if m.creatingQueue {
		var scmd tea.Cmd
		m.spinner, scmd = m.spinner.Update(msg)
		if scmd != nil {
			return m, scmd
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

	// add spinner and toast overlays
	if m.creatingQueue {
		c = m.spinner.View() + " Creating queue...\n" + c
	}
	if m.toast != "" {
		c = fmt.Sprintf("%s\n\n%s", c, m.toast)
	}

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(h + "\n\n" + mainStyle.Render(c) + "\n\n" + f)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}
