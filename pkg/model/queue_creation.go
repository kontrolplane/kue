package model

import (
	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueCreation = "queue creation"

type QueueCreationModel struct {
}

func NewQueueCreationModel() *QueueCreationModel {
	return &QueueCreationModel{}
}

func (m *QueueCreationModel) Init() tea.Cmd {
	return nil
}

func (m *QueueCreationModel) View() string {
	return ""
}

func (m *QueueCreationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
