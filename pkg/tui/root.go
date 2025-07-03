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

	baseCtx := context.Background()
	ctx, cancel := context.WithCancel(baseCtx)

	client, err := client.CreateSqsClient(ctx)
	if err != nil {
		error = fmt.Sprintf("[NewModel] Couldn't create SQS client: %v", err)
	}

    queues, err = kue.ListQueuesUrls(client, ctx)
	if err != nil {
		error = fmt.Sprintf("[NewModel] Error listing queues: %v", err)
	}

	for i, queue := range queues {
        queue, err = kue.FetchQueueAttributes(client, ctx, queue.Url)
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
        context:     ctx,
        cancel:      cancel,
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

    // start polling goroutine
    interval := readRefreshInterval()
    ch := startQueuePolling(ctx, client, interval)
    m.refreshCh = ch

    log.Println("[NewModel] Model initialized")
    return m, nil
}

func (m model) Init() tea.Cmd {
    // kick off first read from refresh channel
    if m.refreshCh != nil {
        return readQueuesCmd(m.refreshCh)
    }
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
