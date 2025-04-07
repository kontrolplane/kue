package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

var (
	dialogBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 0).
			BorderTop(true).
			BorderLeft(true).
			BorderRight(true).
			BorderBottom(true)

	primaryButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("240")).
			Padding(0, 3)

	secondaryButton = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("62")).
			Padding(0, 3)
)

type queueDeleteState struct {
	queue    kue.Queue
	selected int
}

func (m model) QueueDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	return m.SwitchPage(queueDetails), nil
}

func (m model) QueueDeleteView() string {
	question := lipgloss.NewStyle().
		Bold(true).
		Render("Are you sure you want to delete queue: " + m.state.queueDelete.queue.Name + "?")

	confirm := "Yes"
	abort := "No"

	if m.state.queueDelete.selected == 1 {
		confirm = primaryButton.Render(confirm)
		abort = secondaryButton.Render(abort)
	} else {
		confirm = secondaryButton.Render(confirm)
		abort = primaryButton.Render(abort)
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		abort,
		"  ", // Space between buttons
		confirm,
	)

	dialog := lipgloss.JoinVertical(
		lipgloss.Center,
		question,
		"", // Empty line for spacing
		buttons,
	)

	return dialogBox.Render(dialog)
}

func (m model) switchOption() (model, tea.Cmd) {
	m.state.queueDelete.selected = (m.state.queueDelete.selected + 1) % 2
	return m, nil
}

func (m model) QueueDeleteUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Down):
			m, cmd = m.switchOption()
		case key.Matches(msg, m.keys.Up):
			m, cmd = m.switchOption()
		case key.Matches(msg, m.keys.View):
			return m.QueueOverviewSwitchPage(msg)
		case key.Matches(msg, m.keys.Quit):
			m.previous = m.page
			return m.QueueOverviewSwitchPage(msg)
		default:
			m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
		}
	default:
		m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
	}

	return m, cmd
}
