package tui

import (
	"github.com/charmbracelet/huh"

	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueCreate = "queue create"

type queueCreateInput struct {
	queueType              string
	name                   string
	region                 string
	visibilityTimeout      int
	visibilityTimeoutType  string
	messageRetention       int
	messageRetentionType   string
	deliveryDelay          int
	deliveryDelayType      string
	maximumMessageSize     int
	receiveMessageWaitTime int
}

type queueCreateState struct {
	input queueCreateInput
	form  huh.Form
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
	return ""
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
