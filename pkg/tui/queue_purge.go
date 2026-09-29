package tui

import (
	tea "charm.land/bubbletea/v2"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// secondPromptThreshold is the message count above which a purge asks for confirmation twice.
const secondPromptThreshold = 10

// queuePurgeState holds the state for queue purge confirmation.
type queuePurgeState struct {
	queue        kue.Queue
	selected     int  // 0 = no, 1 = yes
	secondPrompt bool // true when showing the second confirmation for large queues
	fromOverview bool // true when triggered from queue overview
}

func (m model) QueuePurgeSwitchPage() (model, tea.Cmd) {
	m.error = ""
	m.state.queuePurge.selected = 0
	m.state.queuePurge.secondPrompt = false
	return m.SwitchPage(queuePurge), nil
}

func (m model) queuePurgeGoBack() (model, tea.Cmd) {
	if m.state.queuePurge.fromOverview {
		return m.QueueOverviewGoBack()
	}
	return m.queueDetailsGoBack()
}

// queuePurgeFinished leaves the purge dialog once the queue is purged, reloading what it returns to.
func (m model) queuePurgeFinished() (model, tea.Cmd) {
	if m.state.queuePurge.fromOverview {
		return m.QueueOverviewGoBack()
	}
	m.state.queueDetails.selectedItems = make(map[string]bool)
	return m.QueueDetailsReload()
}

// queuePurgeMessageCount returns the approximate message count for the purge queue.
func (m model) queuePurgeMessageCount() uint64 {
	return atoi(m.state.queuePurge.queue.ApproximateNumberOfMessages)
}

func (m model) QueuePurgeView() string {
	p := m.state.queuePurge
	name := styles.B(dialogName(p.queue.Name), styles.ToneText)
	note := styles.Faint("sqs allows one purge per queue every 60 seconds.")

	if p.secondPrompt {
		return dialog("purge queue", styles.ToneDanger,
			styles.Render(styles.S("this removes about ", styles.ToneBody),
				styles.B(plural(int(m.queuePurgeMessageCount()), "message"), styles.ToneDanger),
				styles.S(" from ", styles.ToneBody), name, styles.S(", are you sure?", styles.ToneBody)),
			note,
			"",
			confirmButtons(p.selected, styles.ToneDanger),
		)
	}
	return confirmDialog(
		"purge queue",
		[]styles.Span{styles.S("purge all messages from ", styles.ToneBody), name, styles.S("?", styles.ToneBody)},
		[]string{styles.Faint("removes about " + plural(int(m.queuePurgeMessageCount()), "message") + "."), note},
		p.selected,
	)
}

func (m model) QueuePurgeUpdate(msg tea.Msg) (model, tea.Cmd) {
	p := &m.state.queuePurge
	switch m.confirmKey(msg, &p.selected) {
	case confirmNo:
		return m.queuePurgeGoBack()
	case confirmYes:
		if !p.secondPrompt && m.queuePurgeMessageCount() > secondPromptThreshold {
			p.secondPrompt = true
			return m, nil
		}
		m.busy, m.loading, m.loadingMsg = true, true, "purging "+p.queue.Name+"…"
		return m, commands.PurgeQueue(m.context, m.client, p.queue.Name, p.queue.Url)
	}
	return m, nil
}
