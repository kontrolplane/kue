package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueMessageDeleteState holds the state for message deletion confirmation.
type queueMessageDeleteState struct {
	message   kue.Message
	queueUrl  string
	queueName string
	selected  int // 0 = no, 1 = yes
}

func (m model) QueueMessageDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	m.state.queueMessageDelete.selected = 0
	return m.SwitchPage(queueMessageDelete), nil
}

func (m model) QueueMessageDeleteView() string {
	messageID := m.state.queueMessageDelete.message.MessageID
	if len(messageID) > 20 {
		messageID = messageID[:20] + "..."
	}
	messageID = styles.Bold.Render(messageID)
	queueName := styles.Bold.Render(m.state.queueMessageDelete.queueName)

	confirm := "yes"
	abort := "no"

	if m.state.queueMessageDelete.selected == 0 {
		abort = styles.ButtonSecondary.Render(abort)
		confirm = styles.ButtonPrimary.Render(confirm)
	} else {
		abort = styles.ButtonPrimary.Render(abort)
		confirm = styles.ButtonSecondary.Render(confirm)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, abort, "    ", confirm)
	dialog := lipgloss.JoinVertical(lipgloss.Center,
		"warning: message deletion",
		"",
		"are you sure you want to delete message: "+messageID,
		"from queue: "+queueName+" ?",
		"",
		buttons,
	)
	return lipgloss.Place(contentWidth, contentHeight-2, lipgloss.Center, lipgloss.Center, dialog)
}

func (m model) switchMessageDeleteOption() (model, tea.Cmd) {
	m.state.queueMessageDelete.selected = (m.state.queueMessageDelete.selected + 1) % 2
	return m, nil
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Left):
			m, cmd = m.switchMessageDeleteOption()
		case key.Matches(msg, m.keys.Right):
			m, cmd = m.switchMessageDeleteOption()
		case key.Matches(msg, m.keys.View):
			if m.state.queueMessageDelete.selected == 0 {
				if m.previous == queueMessageDetails {
					return m.QueueMessageDetailsSwitchPage(msg)
				}
				return m.QueueDetailsGoBack(msg)
			}
			m.loading = true
			m.loadingMsg = "Deleting message..."
			return m, commands.DeleteMessage(
				m.context,
				m.client,
				m.state.queueMessageDelete.queueUrl,
				m.state.queueMessageDelete.message.ReceiptHandle,
			)
		case key.Matches(msg, m.keys.Quit):
			m.state.queueMessageDelete.selected = 0
			if m.previous == queueMessageDetails {
				return m.QueueMessageDetailsSwitchPage(msg)
			}
			return m.QueueDetailsGoBack(msg)
		}
	}

	return m, cmd
}
