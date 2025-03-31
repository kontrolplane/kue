package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueRedriveState struct {
	queue    kue.Queue
	messages []kue.Message
}

func (m model) QueueRedriveSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueDetails), nil
}

func (m model) QueueRedriveUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m model) QueueRedriveView() string {
	return ""
}
