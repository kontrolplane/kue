package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

type queueMessageDetailsState struct {
	message   kue.Message
	queueName string
	viewport  viewport.Model
	ready     bool
}

func (m model) QueueMessageDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	// Clear any previous error
	m.error = ""

	// Initialize viewport for scrollable content
	m.state.queueMessageDetails.ready = false

	return m.SwitchPage(queueMessageDetails), nil
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Initialize or resize viewport with fixed dimensions
		if !m.state.queueMessageDetails.ready {
			m.state.queueMessageDetails.viewport = viewport.New(contentWidth, contentHeight)
			m.state.queueMessageDetails.viewport.SetContent(m.renderMessageDetails())
			m.state.queueMessageDetails.ready = true
		} else {
			m.state.queueMessageDetails.viewport.Width = contentWidth
			m.state.queueMessageDetails.viewport.Height = contentHeight
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			// Go back to queue details
			return m.QueueDetailsSwitchPage(msg)
		default:
			// Pass to viewport for scrolling
			m.state.queueMessageDetails.viewport, cmd = m.state.queueMessageDetails.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) QueueMessageDetailsView() string {
	if !m.state.queueMessageDetails.ready {
		// Viewport not ready yet, render content directly
		return m.renderMessageDetails()
	}

	// Update viewport content in case it changed
	m.state.queueMessageDetails.viewport.SetContent(m.renderMessageDetails())

	return m.state.queueMessageDetails.viewport.View()
}

// renderMessageDetails renders the full message details content.
func (m model) renderMessageDetails() string {
	msg := m.state.queueMessageDetails.message

	// Styles for the details view
	labelStyle := lipgloss.NewStyle().
		Foreground(styles.AccentColor).
		Bold(true).
		Width(25)

	valueStyle := lipgloss.NewStyle().
		Foreground(styles.TextLight)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.TextWhite).
		Background(styles.AccentColor).
		Padding(0, 1).
		MarginTop(1).
		MarginBottom(1)

	bodyStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(1, 2).
		Width(80)

	// Build the details view
	var sections []string

	// Header with queue name
	header := fmt.Sprintf("Message from queue: %s", m.state.queueMessageDetails.queueName)
	sections = append(sections, lipgloss.NewStyle().Bold(true).Render(header))
	sections = append(sections, "")

	// Basic Information Section
	sections = append(sections, sectionStyle.Render("Basic Information"))

	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Message ID:"),
		valueStyle.Render(msg.MessageID),
	))

	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Sent Timestamp:"),
		valueStyle.Render(msg.SentTimestamp),
	))

	if msg.FirstReceiveTime != "" {
		sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render("First Receive Time:"),
			valueStyle.Render(msg.FirstReceiveTime),
		))
	}

	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Receive Count:"),
		valueStyle.Render(msg.ReceiveCount),
	))

	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("Body Size:"),
		valueStyle.Render(fmt.Sprintf("%d bytes", len(msg.Body))),
	))

	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
		labelStyle.Render("MD5 of Body:"),
		valueStyle.Render(msg.MD5OfBody),
	))

	// FIFO Queue Attributes (if present)
	if msg.MessageGroupID != "" || msg.MessageDeduplicationID != "" || msg.SequenceNumber != "" {
		sections = append(sections, "")
		sections = append(sections, sectionStyle.Render("FIFO Queue Attributes"))

		if msg.MessageGroupID != "" {
			sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render("Message Group ID:"),
				valueStyle.Render(msg.MessageGroupID),
			))
		}

		if msg.MessageDeduplicationID != "" {
			sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render("Deduplication ID:"),
				valueStyle.Render(msg.MessageDeduplicationID),
			))
		}

		if msg.SequenceNumber != "" {
			sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render("Sequence Number:"),
				valueStyle.Render(msg.SequenceNumber),
			))
		}
	}

	// Message Attributes (custom attributes)
	if len(msg.MessageAttributes) > 0 {
		sections = append(sections, "")
		sections = append(sections, sectionStyle.Render("Message Attributes"))

		for attrName, attrValue := range msg.MessageAttributes {
			sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(attrName+":"),
				valueStyle.Render(attrValue),
			))
		}
	}

	// System Attributes
	if len(msg.Attributes) > 0 {
		sections = append(sections, "")
		sections = append(sections, sectionStyle.Render("System Attributes"))

		for attrName, attrValue := range msg.Attributes {
			// Skip attributes we've already shown
			if attrName == "SentTimestamp" || attrName == "ApproximateFirstReceiveTimestamp" ||
				attrName == "ApproximateReceiveCount" || attrName == "MessageGroupId" ||
				attrName == "MessageDeduplicationId" || attrName == "SequenceNumber" {
				continue
			}
			sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(attrName+":"),
				valueStyle.Render(attrValue),
			))
		}
	}

	// Message Body Section
	sections = append(sections, "")
	sections = append(sections, sectionStyle.Render("Message Body"))

	// Truncate body if too long for display
	body := msg.Body
	if len(body) > 2000 {
		body = body[:2000] + "\n... (truncated, showing first 2000 characters)"
	}

	sections = append(sections, bodyStyle.Render(body))

	// Join content left-aligned, then center the whole block
	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	return lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, content)
}
