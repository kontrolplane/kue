package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueDetailsState struct {
	selected int
	queue    kue.Queue
	messages []kue.Message
}

func (m model) QueueDetailsSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueDetails), nil
}

func (m model) NoMessagesFound() bool {
	return m.MessagesCount() == 0
}

func (m model) MessagesCount() int {
	return len(m.state.queueDetails.messages)
}

func (m model) nextMessage() (model, tea.Cmd) {
    if m.state.queueDetails.selected < len(m.state.queueDetails.messages) - 1 {
        m.state.queueDetails.selected++
    }
    return m, nil
}

func (m model) previousMessage() (model, tea.Cmd) {
    if m.state.queueDetails.selected > 0 {
        m.state.queueDetails.selected--
    }
    return m, nil
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) QueueDetailsView() string {
	return ""
}
