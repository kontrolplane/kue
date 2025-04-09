package tui

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
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

	dialog := lipgloss.NewStyle().
		Padding(1, 3)

	secondary := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#628049")).
		Padding(0, 3)

	primary := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("240")).
		Padding(0, 3)

	queueName := lipgloss.NewStyle().
		Bold(true).
		Render(m.state.queueDelete.queue.Name)

	confirm := "yes"
	abort := "no"

	if m.state.queueDelete.selected == 0 {
		confirm = primary.Render(confirm)
		abort = secondary.Render(abort)
	} else {
		confirm = secondary.Render(confirm)
		abort = primary.Render(abort)
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		abort,
		"    ",
		confirm,
	)

	d := lipgloss.JoinVertical(
		lipgloss.Center,
		"warning: queue deletion",
		"",
		"are you sure you want to delete queue: "+queueName+" ?",
		"",
		buttons,
	)

	return dialog.Render(d)
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
				m.state.queueDelete.selected = 0
				return m.QueueOverviewSwitchPage(msg)
			} else {
				if err := kue.DeleteQueue(m.client, m.context, m.state.queueDelete.queue.Name); err != nil {
					m.error = fmt.Sprintf("Error deleting queue: %v", err)
				}
				m.state.queueDelete.selected = 0
				return m.QueueOverviewSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Quit):
			m.state.queueDelete.selected = 0
			return m.QueueOverviewSwitchPage(msg)
		default:
			m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
		}
	default:
		m.state.queueOverview.table, cmd = m.state.queueOverview.table.Update(msg)
	}

	return m, cmd
}
