package tui

import (
	"charm.land/bubbles/v2/key"

	tea "charm.land/bubbletea/v2"
)

var errNoPageSelected = "No page selected"

func (m model) ErrorView() string {
	return m.error
}

func (m model) ErrorUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}
	return m, cmd
}
