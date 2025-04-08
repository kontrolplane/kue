package tui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

var (
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

	log.Println("[QueueDeleteSwitchPage]")

	return m.SwitchPage(queueDelete), nil
}

func (m model) QueueDeleteView() string {

	dialogBox := lipgloss.NewStyle().
		Padding(1, 3)

	question := lipgloss.NewStyle().
		Bold(false).
		Render("are you sure you want to delete queue: " + m.state.queueDelete.queue.Name + "?")

	confirm := "yes"
	abort := "no"

	if m.state.queueDelete.selected == 0 {
		confirm = primaryButton.Render(confirm)
		abort = secondaryButton.Render(abort)
	} else {
		confirm = secondaryButton.Render(confirm)
		abort = primaryButton.Render(abort)
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		abort,
		"    ",
		confirm,
	)

	dialog := lipgloss.JoinVertical(
		lipgloss.Center,
		question,
		"",
		buttons,
	)

	return dialogBox.Render(dialog)
}

func (m model) switchOption() (model, tea.Cmd) {
	m.state.queueDelete.selected = (m.state.queueDelete.selected + 1) % 2
	return m, nil
}

func (m model) QueueDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Left):
			m, cmd = m.switchOption()
		case key.Matches(msg, m.keys.Right):
			m, cmd = m.switchOption()
		case key.Matches(msg, m.keys.View):
			if m.state.queueDelete.selected == 0 {
				m.state.queueDelete.selected = 1
				return m.QueueOverviewSwitchPage(msg)
			}
			if err := kue.DeleteQueue(m.client, m.context, m.state.queueDelete.queue.Name); err != nil {
				m.error = fmt.Sprintf("Error deleting queue: %v", err)
			}
			return m.QueueOverviewSwitchPage(msg)
		case key.Matches(msg, m.keys.Quit):
			return m.QueueOverviewSwitchPage(msg)
		default:
			m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
		}
	default:
		m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
	}

	return m, cmd
}
