package tui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queuePurgeState holds the state for queue purge confirmation.
type queuePurgeState struct {
	queue        kue.Queue
	selected     int  // 0 = no, 1 = yes
	secondPrompt bool // true when showing the second confirmation for large queues
	fromOverview bool // true when triggered from queue overview
}

func (m model) QueuePurgeSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	m.state.queuePurge.selected = 0
	m.state.queuePurge.secondPrompt = false
	return m.SwitchPage(queuePurge), nil
}

func (m model) queuePurgeGoBack(msg tea.Msg) (model, tea.Cmd) {
	if m.state.queuePurge.fromOverview {
		m.error = ""
		return m.SwitchPage(queueOverview), nil
	}
	return m.QueueDetailsGoBack(msg)
}

// queuePurgeMessageCount returns the approximate message count for the purge queue.
func (m model) queuePurgeMessageCount() int {
	n, _ := strconv.Atoi(m.state.queuePurge.queue.ApproximateNumberOfMessages)
	return n
}

func (m model) QueuePurgeView() string {
	queueDisplay := styles.Bold.Render(m.state.queuePurge.queue.Name)

	confirm := "yes"
	abort := "no"

	if m.state.queuePurge.selected == 0 {
		abort = styles.ButtonSecondary.Render(abort)
		confirm = styles.ButtonPrimary.Render(confirm)
	} else {
		abort = styles.ButtonPrimary.Render(abort)
		confirm = styles.ButtonSecondary.Render(confirm)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, abort, "    ", confirm)

	var prompt string
	if m.state.queuePurge.secondPrompt {
		count := m.queuePurgeMessageCount()
		dangerStyle := lipgloss.NewStyle().Foreground(styles.DangerRed).Bold(true)
		prompt = fmt.Sprintf(
			"this queue has %s messages, are you really sure?",
			dangerStyle.Render(fmt.Sprintf("~%d", count)),
		)
	} else {
		prompt = "are you sure you want to purge all messages from: " + queueDisplay + " ?"
	}

	dialog := lipgloss.JoinVertical(lipgloss.Center,
		"warning: queue purge",
		"",
		prompt,
		"",
		buttons,
	)
	return lipgloss.Place(contentWidth, contentHeight-2, lipgloss.Center, lipgloss.Center, dialog)
}

func (m model) switchPurgeOption() (model, tea.Cmd) {
	m.state.queuePurge.selected = (m.state.queuePurge.selected + 1) % 2
	return m, nil
}

func (m model) QueuePurgeUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Left):
			m, cmd = m.switchPurgeOption()
		case key.Matches(msg, m.keys.Right):
			m, cmd = m.switchPurgeOption()
		case key.Matches(msg, m.keys.View):
			if m.state.queuePurge.selected == 0 {
				return m.queuePurgeGoBack(msg)
			}
			// If >10 messages and haven't shown second prompt yet, show it
			if !m.state.queuePurge.secondPrompt && m.queuePurgeMessageCount() > 10 {
				m.state.queuePurge.secondPrompt = true
				m.state.queuePurge.selected = 0
				return m, nil
			}
			m.loading = true
			m.loadingMsg = "Purging queue..."
			return m, commands.PurgeQueue(m.context, m.client, m.state.queuePurge.queue.Url)
		case key.Matches(msg, m.keys.Quit):
			m.state.queuePurge.selected = 0
			return m.queuePurgeGoBack(msg)
		}
	}

	return m, cmd
}
