package model

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
	keys "github.com/kontrolplane/kue/pkg/keys"
)

type MainModel struct {
	projectName string
	programName string
	viewName    string

	width  int
	height int

	keys keys.KeyMap
	help help.Model

	model tea.Model
}

var (
	projectName = "kontrolplane"
	programName = "kue"
)

func NewMainModel() MainModel {

	return MainModel{
		projectName: projectName,
		programName: programName,
		viewName:    viewNameQueueOverview,

		keys: keys.Keys,
		help: help.NewModel(),

		model: NewQueueOverviewModel(),
	}
}

func (m MainModel) Init() tea.Cmd {
	return nil
}

func (m MainModel) View() string {
	return ""
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	return m, nil
}
