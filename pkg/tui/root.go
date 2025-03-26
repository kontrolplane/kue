package tui

import (
	"context"
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
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
		log.Printf("[NewModel] fetching queue attributes: %s", queue.Url)

		queue, err = kue.FetchQueueAttributes(client, context, queue.Url)
		if err != nil {
			error = fmt.Sprintf("[NewModel] Error fetching queue attributes: %v", err)
		}

		log.Printf("[NewModel] fetched queue attributes: %s", queue)
		queues[i] = queue
	}

	log.Println("[NewModel] Initializing new model")
	queueOverviewTable := initQueueOverviewTable()

	m := model{
		projectName: projectName,
		programName: programName,
		viewName:    viewNameQueueOverview,

		page:    queueOverview,
		context: context,

		error: error,

		keys: keys.Keys,
		help: help.New(),

		state: state{
			queueOverview: queueOverviewState{
				selected: 0,
				table:    queueOverviewTable,
				queues:   queues,
			},
			queueDetails: queueDetailsState{},
		},
	}

	log.Println("[NewModel] Model initialized")
	return m, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("[Update] Received message of type: %T", msg)

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		log.Printf("[Update] Window size changed to %dx%d", msg.Width, msg.Height)
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		log.Printf("[Update] Key pressed: %s", msg.String())
		switch {
		case key.Matches(msg, m.keys.Select):
			log.Println("[Update] Select key pressed")
		case key.Matches(msg, m.keys.View):
			log.Println("[Update] View key pressed")
		case key.Matches(msg, m.keys.Help):
			log.Println("[Update] Help key pressed")
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Quit):
			log.Println("[Update] Quit key pressed")
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	switch m.page {
	case queueOverview:
		m, cmd = m.QueueOverviewUpdate(msg)
	}

	return m, cmd
}

func (m model) View() string {
	log.Printf("[View] Rendering view for page: %d, queue count: %d", m.page, len(m.state.queueOverview.queues))

	var h string = formatHeader(m.projectName, m.programName, m.viewName)
	var f string = m.help.View(m.keys)
	var c string

	switch m.page {
	case queueOverview:
		c = m.QueueOverviewView()
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
