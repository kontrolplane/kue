package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

var viewNameQueueDetails = "queue details"

type queueDetailsState struct {
	selected int
	queue    kue.Queue
	messages []kue.Message
}

func (m model) QueueDetailsSwitch(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueDetails), nil
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {

	return m, nil
}

func (m model) QueueDetailsView() string {
	return ""
}
