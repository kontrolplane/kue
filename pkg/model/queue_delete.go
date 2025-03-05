package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueDelete = "queue delete"

type QueueDeleteModel struct {
}

func NewQueueDeleteModel() *QueueDeleteModel {
	return &QueueDeleteModel{}
}

func (m *QueueDeleteModel) Init() tea.Cmd {
	return nil
}

func (m *QueueDeleteModel) View() string {
	return ""
}

func (m *QueueDeleteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
