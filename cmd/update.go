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
		m.width = msg.Width
		m.height = msg.Height

	// Handle key presses, these are shown at the bottom of the view.
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
		case key.Matches(msg, m.keys.Down):
		case key.Matches(msg, m.keys.Left):
		case key.Matches(msg, m.keys.Right):
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
