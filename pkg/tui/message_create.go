package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

var viewNameQueueMessageCreate = "queue message create"

type queueMessageCreateInput struct {
	messageBody   string
	deliveryDelay string
}

type queueMessageCreateState struct {
	input queueMessageCreateInput
	form  huh.Form
}

func (m model) QueueMessageCreateSwitchPage(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.SwitchPage(queueMessageCreate), nil
}

func (m model) QueueMessageCreateView() string {
	return ""
}

func (m model) QueueMessageCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
