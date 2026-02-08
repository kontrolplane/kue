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

const (
	detailsRightPanelWidth   = 74
	detailsRightContentWidth = detailsRightPanelWidth - 4
	detailsViewportHeight    = contentHeight - 3 // Account for header and margin
)

// queueMessageDetailsState holds the state for message details view.
type queueMessageDetailsState struct {
	message   kue.Message
	queueName string
	queueUrl  string
	isFifo    bool
	viewport  viewport.Model
}

func (m model) QueueMessageDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""

	// Initialize viewport for message body
	vp := viewport.New(detailsRightContentWidth, detailsViewportHeight)
	vp.SetContent(m.state.queueMessageDetails.message.Body)
	m.state.queueMessageDetails.viewport = vp

	return m.SwitchPage(queueMessageDetails), nil
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.DeleteMessage):
			if m.state.queueMessageDetails.message.ReceiptHandle != "" {
				m.state.queueMessageDelete.message = m.state.queueMessageDetails.message
				m.state.queueMessageDelete.queueUrl = m.state.queueMessageDetails.queueUrl
				m.state.queueMessageDelete.queueName = m.state.queueMessageDetails.queueName
				return m.QueueMessageDeleteSwitchPage(msg)
			}
		case key.Matches(msg, m.keys.Quit):
			return m.QueueDetailsGoBack(msg)
		}
	}

	// Update viewport for scrolling
	m.state.queueMessageDetails.viewport, cmd = m.state.queueMessageDetails.viewport.Update(msg)
	return m, cmd
}

func (m model) QueueMessageDetailsView() string {
	return m.renderMessageDetails()
}

func (m model) renderMessageDetails() string {
	msg := m.state.queueMessageDetails.message

	const (
		leftPanelWidth   = 62
		leftContentWidth = leftPanelWidth - 4 // Account for padding (4)
		labelWidth       = 16
		valueWidth       = leftContentWidth - labelWidth - 2
	)

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
		Width(leftContentWidth)

	row := func(label, value string) string {
		return lipgloss.JoinHorizontal(lipgloss.Top,
			labelStyle.Render(label),
			valueStyle.Render(value),
		)
	}

	// Left panel - queue information first
	var leftSections []string
	leftSections = append(leftSections, sectionHeader.Render("Queue Information"))
	leftSections = append(leftSections, row("Queue Name", m.state.queueMessageDetails.queueName))
	queueType := "Standard"
	if m.state.queueMessageDetails.isFifo {
		queueType = "FIFO"
	}
	leftSections = append(leftSections, row("Queue Type", queueType))

	// Message metadata
	leftSections = append(leftSections, sectionHeader.MarginTop(1).Render("Basic Information"))
	leftSections = append(leftSections, row("Message ID", msg.MessageID))
	leftSections = append(leftSections, row("Sent", msg.SentTimestamp))
	if msg.FirstReceiveTime != "" {
		leftSections = append(leftSections, row("First Received", msg.FirstReceiveTime))
	}
	leftSections = append(leftSections, row("Receive Count", msg.ReceiveCount))
	leftSections = append(leftSections, row("Body Size", fmt.Sprintf("%d bytes", len(msg.Body))))
	leftSections = append(leftSections, row("MD5", msg.MD5OfBody))

	if msg.MessageGroupID != "" || msg.MessageDeduplicationID != "" || msg.SequenceNumber != "" {
		leftSections = append(leftSections, sectionHeader.MarginTop(1).Render("FIFO Attributes"))
		if msg.MessageGroupID != "" {
			leftSections = append(leftSections, row("Group ID", msg.MessageGroupID))
		}
		if msg.MessageDeduplicationID != "" {
			leftSections = append(leftSections, row("Dedup ID", msg.MessageDeduplicationID))
		}
		if msg.SequenceNumber != "" {
			leftSections = append(leftSections, row("Sequence", msg.SequenceNumber))
		}
	}

	if len(msg.MessageAttributes) > 0 {
		leftSections = append(leftSections, sectionHeader.MarginTop(1).Render("Custom Attributes"))
		for name, value := range msg.MessageAttributes {
			leftSections = append(leftSections, row(name, value))
		}
	}

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
		leftSections = append(leftSections, sectionHeader.MarginTop(1).Render("System Attributes"))
		leftSections = append(leftSections, sysAttrs...)
	}

	leftPanelStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2).
		Width(leftPanelWidth).
		Height(contentHeight)

	leftPanel := leftPanelStyle.Render(lipgloss.JoinVertical(lipgloss.Left, leftSections...))

	// Vertical divider - create full height line
	var dividerLines string
	for i := 0; i < contentHeight; i++ {
		dividerLines += "│"
		if i < contentHeight-1 {
			dividerLines += "\n"
		}
	}
	dividerStyle := lipgloss.NewStyle().
		Foreground(styles.BorderColor)

	divider := dividerStyle.Render(dividerLines)

	// Right panel - message body with viewport
	bodyHeaderStyle := lipgloss.NewStyle().
		Foreground(styles.TextLight).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(styles.BorderColor).
		Width(detailsRightContentWidth).
		MarginBottom(1)

	rightPanelStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2).
		Width(detailsRightPanelWidth).
		Height(contentHeight)

	rightContent := lipgloss.JoinVertical(lipgloss.Left,
		bodyHeaderStyle.Render("Message Body"),
		m.state.queueMessageDetails.viewport.View(),
	)
	rightPanel := rightPanelStyle.Render(rightContent)

	// Join panels horizontally with divider
	content := lipgloss.JoinHorizontal(lipgloss.Top,
		leftPanel,
		divider,
		rightPanel,
	)

	return lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, content)
}
