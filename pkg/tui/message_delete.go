package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueMessageDeleteState holds the state for message deletion confirmation.
type queueMessageDeleteState struct {
	messages    []kue.Message
	queueUrl    string
	queueName   string
	selected    int  // 0 = no, 1 = yes
	fromDetails bool // true when triggered from the message details view
}

func (m model) QueueMessageDeleteSwitchPage() (model, tea.Cmd) {
	m.error = ""
	m.state.queueMessageDelete.selected = 0
	return m.SwitchPage(queueMessageDelete), nil
}

func (m model) messageDeleteGoBack() (model, tea.Cmd) {
	if m.state.queueMessageDelete.fromDetails {
		m.error = ""
		return m.SwitchPage(queueMessageDetails), nil
	}
	return m.queueDetailsGoBack()
}

func (m model) QueueMessageDeleteView() string {
	d := m.state.queueMessageDelete
	ids := make([]string, len(d.messages))
	for i, msg := range d.messages {
		ids[i] = msg.MessageID
	}

	target := styles.B(plural(len(ids), "message"), styles.ToneText)
	if len(ids) == 1 {
		target = styles.B("message "+dialogName(ids[0]), styles.ToneText)
	}
	body := listBody(ids, false)
	body = append(body, "", styles.Faint("deleted messages cannot be brought back."))
	return confirmDialog(
		"delete message",
		[]styles.Span{styles.S("delete ", styles.ToneBody), target, styles.S(" from ", styles.ToneBody),
			styles.B(dialogName(d.queueName), styles.ToneText), styles.S("?", styles.ToneBody)},
		body,
		d.selected,
	)
}

func (m model) QueueMessageDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
	d := &m.state.queueMessageDelete
	switch m.confirmKey(msg, &d.selected) {
	case confirmNo:
		return m.messageDeleteGoBack()
	case confirmYes:
		m.busy, m.loading, m.loadingMsg = true, true, deletingMsg(len(d.messages), "message")
		return m, commands.DeleteMessages(m.context, m.client, d.queueUrl, d.messages)
	}
	return m, nil
}
