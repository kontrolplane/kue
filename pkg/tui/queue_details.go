package tui

import (
	"fmt"
	"log"
	"reflect"
	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueDetailsState struct {
	selected int
	queue    kue.Queue
	table    table.Model
}

var attributeColumns []table.Column = []table.Column{
	{
		Title: "attribute", Width: 40,
	},
	{
		Title: "value", Width: 60,
	},
}

func initAttributeTable() table.Model {
	t := table.New(
		table.WithColumns(attributeColumns),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#628049")).
		Bold(false)

	t.SetStyles(s)
	return t
}

// buildQueueAttributeRows converts a Queue struct (including tags) into rows for the attribute table.
func buildQueueAttributeRows(q kue.Queue) []table.Row {
	var rows []table.Row

	v := reflect.ValueOf(q)
	t := reflect.TypeOf(q)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		valueField := v.Field(i)

		// Skip zero-value fields to avoid showing empty attributes
		val := fmt.Sprintf("%v", valueField.Interface())
		if val == "" || val == "map[]" {
			continue
		}

		// Skip Tags field; we'll handle separately
		if field.Name == "Tags" {
			continue
		}

		rows = append(rows, table.Row{field.Name, val})
	}

	// Handle tags (sorted for deterministic order)
	if len(q.Tags) > 0 {
		var keys []string
		for k := range q.Tags {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			rows = append(rows, table.Row{"tag:" + k, q.Tags[k]})
		}
	}

	return rows
}

func (m model) QueueDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {

	log.Println("[QueueDetailsSwitchPage]")

	// Refresh the queue attributes in case they have changed.
	queue, err := kue.FetchQueueAttributes(m.client, m.context, m.state.queueDetails.queue.Url)
	if err != nil {
		m.error = fmt.Sprintf("Error fetching queue attributes: %v", err)
	} else {
		m.state.queueDetails.queue = queue
	}

	rows := buildQueueAttributeRows(m.state.queueDetails.queue)
	m.state.queueDetails.table.SetRows(rows)
	m.state.queueDetails.table.SetCursor(0)

	return m.SwitchPage(queueDetails), nil
}

func (m model) QueueDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m.QueueOverviewSwitchPage(msg)
		default:
			m.state.queueDetails.table, cmd = m.state.queueDetails.table.Update(msg)
		}
	default:
		m.state.queueDetails.table, cmd = m.state.queueDetails.table.Update(msg)
	}

	return m, cmd
}

func (m model) QueueDetailsView() string {
	log.Println("[QueueDetailsView] queue:", m.state.queueDetails.queue.Name)
	return m.state.queueDetails.table.View()
}
