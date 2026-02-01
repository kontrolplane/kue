package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

type queueDeleteState struct {
	queue    kue.Queue
	selected int
}

func (m model) QueueDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	// Clear any previous error
	m.error = ""
	return m.SwitchPage(queueDelete), nil
}

func (m model) QueueDeleteView() string {
	queueName := styles.Bold.Render(m.state.queueDelete.queue.Name)

	confirm := "yes"
	abort := "no"

	if m.state.queueDelete.selected == 0 {
		confirm = styles.ButtonPrimary.Render(confirm)
		abort = styles.ButtonSecondary.Render(abort)
	} else {
		confirm = styles.ButtonSecondary.Render(confirm)
		abort = styles.ButtonPrimary.Render(abort)
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

	return styles.DialogContainer.Render(d)
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
				// User selected "no" - go back
				m.state.queueDelete.selected = 0
				return m.QueueOverviewSwitchPage(msg)
			} else {
				// User selected "yes" - delete the queue async
				m.loading = true
				m.loadingMsg = "Deleting queue..."
				return m, commands.DeleteQueue(m.context, m.client, m.state.queueDelete.queue.Name)
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
