package model

import (
	"github.com/charmbracelet/bubbles/list"

	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueDetails = "queue details"

type QueueDetailsModel struct {
	list list.Model
}

func NewQueueDetailsModel() *QueueDetailsModel {
	return &QueueDetailsModel{}
}

func (m *QueueDetailsModel) Init() tea.Cmd {
	return nil
}

func (m *QueueDetailsModel) View() string {
	return ""
}

func (m *QueueDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
