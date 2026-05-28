package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"

	tea "charm.land/bubbletea/v2"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueDeleteState holds the state for queue deletion confirmation.
type queueDeleteState struct {
	queues   []kue.Queue
	selected int // 0 = no, 1 = yes
}

func (m model) QueueDeleteSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	m.state.queueDelete.selected = 0
	return m.SwitchPage(queueDelete), nil
}

func (m model) QueueDeleteView() string {
	numQueues := len(m.state.queueDelete.queues)

	var prompt string
	if numQueues == 1 {
		prompt = "are you sure you want to delete: " + styles.Bold.Render(m.state.queueDelete.queues[0].Name) + " ?"
	} else {
		prompt = fmt.Sprintf("are you sure you want to delete: %s ?", styles.Bold.Render(fmt.Sprintf("%d queues", numQueues)))
	}

	return renderConfirmDialog("warning: queue deletion", m.state.queueDelete.selected, prompt)
}

func (m model) QueueDeleteUpdate(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.Left), key.Matches(msg, m.keys.Right):
			m.state.queueDelete.selected = (m.state.queueDelete.selected + 1) % 2
		case key.Matches(msg, m.keys.View):
			if m.state.queueDelete.selected == 0 {
				return m.QueueOverviewSwitchPage(msg)
			}
			m.loading = true
			numQueues := len(m.state.queueDelete.queues)
			if numQueues == 1 {
				m.loadingMsg = "Deleting queue..."
				return m, commands.DeleteQueue(m.context, m.client, m.state.queueDelete.queues[0].Name)
			}
			m.loadingMsg = fmt.Sprintf("Deleting %d queues...", numQueues)
			return m, commands.DeleteQueues(m.context, m.client, m.state.queueDelete.queues)
		case key.Matches(msg, m.keys.Quit):
			m.state.queueDelete.selected = 0
			return m.QueueOverviewSwitchPage(msg)
		}
	}

	return m, nil
}
