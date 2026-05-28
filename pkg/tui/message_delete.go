package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"

	tea "charm.land/bubbletea/v2"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueMessageDeleteState holds the state for message deletion confirmation.
type queueMessageDeleteState struct {
	messages  []kue.Message
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
	numMessages := len(m.state.queueMessageDelete.messages)

	var messageDisplay string
	if numMessages == 1 {
		messageID := m.state.queueMessageDelete.messages[0].MessageID
		if len(messageID) > 20 {
			messageID = messageID[:20] + "..."
		}
		messageDisplay = styles.Bold.Render(messageID)
	} else {
		messageDisplay = styles.Bold.Render(fmt.Sprintf("%d messages", numMessages))
	}
	queueName := styles.Bold.Render(m.state.queueMessageDelete.queueName)

	return renderConfirmDialog(
		"warning: message deletion",
		m.state.queueMessageDelete.selected,
		"are you sure you want to delete: "+messageDisplay,
		"from queue: "+queueName+" ?",
	)
}

func (m model) queueMessageDeleteGoBack(msg tea.Msg) (model, tea.Cmd) {
	if m.previous == queueMessageDetails {
		return m.QueueMessageDetailsSwitchPage(msg)
	}
	return m.QueueDetailsGoBack(msg)
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Left), key.Matches(msg, m.keys.Right):
			m.state.queueMessageDelete.selected = (m.state.queueMessageDelete.selected + 1) % 2
		case key.Matches(msg, m.keys.View):
			if m.state.queueMessageDelete.selected == 0 {
				return m.queueMessageDeleteGoBack(msg)
			}
			m.loading = true
			numMessages := len(m.state.queueMessageDelete.messages)
			if numMessages == 1 {
				m.loadingMsg = "Deleting message..."
				return m, commands.DeleteMessage(
					m.context,
					m.client,
					m.state.queueMessageDelete.queueUrl,
					m.state.queueMessageDelete.messages[0].ReceiptHandle,
				)
			}
			m.loadingMsg = fmt.Sprintf("Deleting %d messages...", numMessages)
			return m, commands.DeleteMessages(
				m.context,
				m.client,
				m.state.queueMessageDelete.queueUrl,
				m.state.queueMessageDelete.messages,
			)
		case key.Matches(msg, m.keys.Quit):
			m.state.queueMessageDelete.selected = 0
			return m.queueMessageDeleteGoBack(msg)
		}
	}

	return m, nil
}
