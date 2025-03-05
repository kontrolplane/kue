package model

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
)

var viewNameQueueOverview = "queue overview"

var columnMap = map[int]string{
	0: "account",
	1: "service endpoint",
	2: "queue name",
}

var queueOverviewColumns []table.Column = []table.Column{
	{
		Title: columnMap[0], Width: 20,
	},
	{
		Title: columnMap[1], Width: 50,
	},
	{
		Title: columnMap[2], Width: 40,
	},
}

var queueOverviewRows []table.Row = []table.Row{}

type QueueOverviewModel struct {
	table table.Model
}

func NewQueueOverviewModel() *QueueOverviewModel {

	/**
	for _, q := range queues {
		queueOverviewRows = append(queueOverviewRows, table.Row{
			q.AccountIdentifier,
			q.ServiceEndpoint,
			q.Name,
		})
	}
	*/

	t := table.New(
		table.WithColumns(queueOverviewColumns),
		table.WithRows(queueOverviewRows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#628049")).
		Bold(false)

	t.SetStyles(s)

	return &QueueOverviewModel{
		table: t,
	}
}

func (m QueueOverviewModel) Init() tea.Cmd {
	return nil
}

func (m QueueOverviewModel) View() string {
	return ""
}

func (m QueueOverviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}
