package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueDeleteState holds the state for queue deletion confirmation.
type queueDeleteState struct {
	queue    kue.Queue
	selected int // 0 = no, 1 = yes
}

func (m model) QueueDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	m.state.queueDelete.selected = 0
	return m.SwitchPage(queueDelete), nil
}

func (m model) QueueDeleteView() string {
	queueName := styles.Bold.Render(m.state.queueDelete.queue.Name)

	confirm := "yes"
	abort := "no"

	if m.state.queueDelete.selected == 0 {
		abort = styles.ButtonSecondary.Render(abort)
		confirm = styles.ButtonPrimary.Render(confirm)
	} else {
		abort = styles.ButtonPrimary.Render(abort)
		confirm = styles.ButtonSecondary.Render(confirm)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, abort, "    ", confirm)
	dialog := lipgloss.JoinVertical(lipgloss.Center,
		"warning: queue deletion",
		"",
		"are you sure you want to delete queue: "+queueName+" ?",
		"",
		buttons,
	)
	return lipgloss.Place(contentWidth, contentHeight-2, lipgloss.Center, lipgloss.Center, dialog)
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
				return m.QueueOverviewSwitchPage(msg)
			}
			m.loading = true
			m.loadingMsg = "Deleting queue..."
			return m, commands.DeleteQueue(m.context, m.client, m.state.queueDelete.queue.Name)
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
