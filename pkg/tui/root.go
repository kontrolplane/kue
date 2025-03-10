package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	keys "github.com/kontrolplane/kue/pkg/keys"
	sqs "github.com/kontrolplane/kue/pkg/sqs"
)

type page uint

const (
	queueOverview page = iota
	queueDetails
	queueCreate
	queueDelete
	queueMessageDetails
	queueMessageCreate
	queuemessageDelete
)

type model struct {
	projectName string
	programName string
	viewName    string

	page     page
	previous page

	queues []sqs.Queue

	context context.Context

	width  int
	height int

	keys keys.KeyMap
	help help.Model

	table tea.Model
}

var (
	projectName = "kontrolplane"
	programName = "kue"
)

var mainStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func NewModel(
	projectName string,
	programName string,
) (tea.Model, error) {

	context := context.Background()

	return model{
		projectName: projectName,
		programName: programName,
		viewName:    viewNameQueueOverview,

		page:    queueOverview,
		context: context,

		keys: keys.Keys,
		help: help.New(),
	}, nil
}

func (m model) SwitchPage(page page) model {
	m.page = page
	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	var h string = fmt.Sprintf("%s/%s • %s", m.projectName, m.programName, m.viewName)
	var f string = m.help.View(m.keys)

	content := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center).
		Render(h + "\n\n" + mainStyle.Render(m.table.View()) + "\n\n" + f)

	return lipgloss.JoinVertical(lipgloss.Top, content)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg := msg.(type) {

	// Handle window resizes by updating the width and height in the model.
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	// Handle key presses, these are shown at the bottom of the view.
	case tea.KeyMsg:
		switch {

		case key.Matches(msg, m.keys.Select):
		case key.Matches(msg, m.keys.View):

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	switch m.page {
	case queueOverview:
		// m, cmd = m.QueueOverviewUpdate(msg)
	}

	return m, cmd
}
