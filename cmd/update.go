package cmd

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	// Handle window resizes by updating the width and height in the model.
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

	// Handle key presses, these are shown at the bottom of the view.
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.Keys.Up):
		case key.Matches(msg, m.Keys.Down):
		case key.Matches(msg, m.Keys.Left):
		case key.Matches(msg, m.Keys.Right):
		case key.Matches(msg, m.Keys.Select):
		case key.Matches(msg, m.Keys.View):

		case key.Matches(msg, m.Keys.Help):
			m.Help.ShowAll = !m.Help.ShowAll
		case key.Matches(msg, m.Keys.Quit):
			return m, tea.Quit
		}
	}

	return m, nil
}
