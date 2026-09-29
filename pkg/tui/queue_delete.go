package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueDeleteState holds the state for queue deletion confirmation.
type queueDeleteState struct {
	queues  []string
	hidden  int // selected queues the overview filter hides, which are not deleted
	confirm typedConfirm
}

// QueueDeleteSwitchPage asks to type the name of the queue, or how many are deleted, since a
// deleted queue cannot be brought back.
func (m model) QueueDeleteSwitchPage() (model, tea.Cmd) {
	m.error = ""
	d := &m.state.queueDelete
	want := fmt.Sprintf("delete %d queues", len(d.queues))
	if len(d.queues) == 1 {
		want = d.queues[0]
	}
	var cmd tea.Cmd
	d.confirm, cmd = newTypedConfirm(want)
	return m.SwitchPage(queueDelete), cmd
}

func (m model) QueueDeleteView() string {
	d := m.state.queueDelete

	target, their := styles.B(plural(len(d.queues), "queue"), styles.ToneText), "their"
	if len(d.queues) == 1 {
		target, their = styles.B(dialogName(d.queues[0]), styles.ToneText), "its"
	}
	body := []string{styles.Render(styles.S("delete ", styles.ToneBody), target, styles.S(" with all of "+their+" messages?", styles.ToneBody))}
	body = append(body, listBody(d.queues, false)...)
	body = append(body, "", styles.Faint("this cannot be undone, a queue with the same name can only be created again after 60 seconds."))
	if d.hidden > 0 {
		body = append(body, styles.Faint(fmt.Sprintf("%s hidden by the filter %s not included.", plural(d.hidden, "selected queue"), isAre(d.hidden))))
	}
	body = append(body, "", d.confirm.view())
	return dialog("delete queue", styles.ToneDanger, body...)
}

func isAre(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}

func (m model) QueueDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
	d := &m.state.queueDelete
	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Back):
			return m.QueueOverviewGoBack()
		case key.Matches(keyMsg, m.keys.View):
			if !d.confirm.ok() {
				return m, nil
			}
			m.busy, m.loading, m.loadingMsg = true, true, deletingMsg(len(d.queues), "queue")
			return m, commands.DeleteQueues(m.context, m.client, d.queues)
		}
	}
	var cmd tea.Cmd
	d.confirm, cmd = d.confirm.update(msg)
	return m, cmd
}
