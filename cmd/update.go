package cmd

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// Update is called when a message is received. Use it to inspect messages
// and, in response, update the model and/or send a command.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	var c tea.Cmd

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
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}

				if m.state == QueueOverview {
					m.state = QueueDetails
				}
			}

		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll

		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	m.table, c = m.table.Update(msg)
	return m, c
}
