package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

type queueMessageDetailsState struct {
	message   kue.Message
	queueName string
	queueUrl  string
}

func (m model) QueueMessageDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	// Clear any previous error
	m.error = ""

	return m.SwitchPage(queueMessageDetails), nil
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.DeleteMessage):
			// Navigate to message delete confirmation
			if m.state.queueMessageDetails.message.ReceiptHandle != "" {
				m.state.queueMessageDelete.message = m.state.queueMessageDetails.message
				m.state.queueMessageDelete.queueUrl = m.state.queueMessageDetails.queueUrl
				m.state.queueMessageDelete.queueName = m.state.queueMessageDetails.queueName
				return m.QueueMessageDeleteSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Quit):
			// Go back to queue details without reloading
			return m.QueueDetailsGoBack(msg)
		}
	}

	return m, nil
}

func (m model) QueueMessageDetailsView() string {
	return m.renderMessageDetails()
}

// renderMessageDetails renders the full message details content.
func (m model) renderMessageDetails() string {
	msg := m.state.queueMessageDetails.message

	// Define consistent widths
	const (
		labelWidth   = 22
		valueWidth   = 45
		sectionWidth = labelWidth + valueWidth + 3
	)

	// Reusable styles
	labelStyle := lipgloss.NewStyle().
		Foreground(styles.MediumGray).
		Width(labelWidth).
		Align(lipgloss.Right).
		PaddingRight(2)

	valueStyle := lipgloss.NewStyle().
		Foreground(styles.TextLight).
		Width(valueWidth)

	sectionHeader := lipgloss.NewStyle().
		Foreground(styles.TextLight).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(styles.BorderColor).
		Width(sectionWidth).
		MarginTop(1)

	rowStyle := lipgloss.NewStyle().PaddingLeft(2)

	bodyBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1, 2).
		Width(sectionWidth).
		MarginTop(1)

	// Helper to create a row
	row := func(label, value string) string {
		return rowStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(label),
				valueStyle.Render(value),
			),
		)
	}

	var sections []string

	// Basic Information
	sections = append(sections, sectionHeader.Render("Basic Information"))
	sections = append(sections, row("Message ID", msg.MessageID))
	sections = append(sections, row("Sent", msg.SentTimestamp))
	if msg.FirstReceiveTime != "" {
		sections = append(sections, row("First Received", msg.FirstReceiveTime))
	}
	sections = append(sections, row("Receive Count", msg.ReceiveCount))
	sections = append(sections, row("Body Size", fmt.Sprintf("%d bytes", len(msg.Body))))
	sections = append(sections, row("MD5", msg.MD5OfBody))

	// FIFO Queue Attributes (if present)
	if msg.MessageGroupID != "" || msg.MessageDeduplicationID != "" || msg.SequenceNumber != "" {
		sections = append(sections, sectionHeader.Render("FIFO Attributes"))
		if msg.MessageGroupID != "" {
			sections = append(sections, row("Group ID", msg.MessageGroupID))
		}
		if msg.MessageDeduplicationID != "" {
			sections = append(sections, row("Deduplication ID", msg.MessageDeduplicationID))
		}
		if msg.SequenceNumber != "" {
			sections = append(sections, row("Sequence Number", msg.SequenceNumber))
		}
	}

	// Message Attributes (custom)
	if len(msg.MessageAttributes) > 0 {
		sections = append(sections, sectionHeader.Render("Custom Attributes"))
		for name, value := range msg.MessageAttributes {
			sections = append(sections, row(name, value))
		}
	}

	// System Attributes (excluding already shown)
	var sysAttrs []string
	skip := map[string]bool{
		"SentTimestamp": true, "ApproximateFirstReceiveTimestamp": true,
		"ApproximateReceiveCount": true, "MessageGroupId": true,
		"MessageDeduplicationId": true, "SequenceNumber": true,
	}
	for name, value := range msg.Attributes {
		if !skip[name] {
			sysAttrs = append(sysAttrs, row(name, value))
		}
	}
	if len(sysAttrs) > 0 {
		sections = append(sections, sectionHeader.Render("System Attributes"))
		sections = append(sections, sysAttrs...)
	}

	// Message Body
	sections = append(sections, sectionHeader.Render("Body"))
	body := msg.Body
	if len(body) > 2000 {
		body = body[:2000] + "\n... (truncated)"
	}
	sections = append(sections, bodyBoxStyle.Render(body))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, content)
}
