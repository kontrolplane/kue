package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	sqs "github.com/kontrolplane/kue/pkg/sqs"
)

var viewNameQueueDetails = "queue details"

type queueDetailsState struct {
	selected int
	queue    sqs.Queue
	messages []sqs.Message
}

func (m model) QueueDetailsView() string {
	return ""
}

func (m model) QueueDetailsSwitch(msg tea.Msg) (tea.Model, tea.Cmd) {
	m = m.SwitchPage(queueDetails)
	return m, nil
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	m = m.SwitchPage(queueDetails)

	return m, nil
}
