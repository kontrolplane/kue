package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueMessageCreateState holds the state for message creation.
type queueMessageCreateState struct {
	queueName string
	queueUrl  string
	isFifo    bool
	textarea  textarea.Model
	selected  int // 0 = textarea, 1 = cancel, 2 = submit
}

const (
	messageLeftPanelWidth  = 50
	messageRightPanelWidth = 86
	messageTextareaWidth   = messageRightPanelWidth - 4 // Panel padding(4)
	messageTextareaHeight  = contentHeight - 8          // Height for textarea minus headers and panel padding
)

func (m model) QueueMessageCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""

	ta := textarea.New()
	ta.Placeholder = "Enter message body (JSON or plain text)..."
	ta.Focus()
	ta.SetWidth(messageTextareaWidth)
	ta.SetHeight(messageTextareaHeight)
	ta.CharLimit = 262144

	m.state.queueMessageCreate.textarea = ta
	m.state.queueMessageCreate.selected = 0
	return m.SwitchPage(queueMessageCreate), nil
}

func (m model) QueueMessageCreateView() string {
	const (
		leftContentWidth  = messageLeftPanelWidth - 4  // Account for padding (4)
		rightContentWidth = messageRightPanelWidth - 4 // Account for padding (4)
		labelWidth        = 14
		valueWidth        = leftContentWidth - labelWidth - 2
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

	// Left panel - top section (queue information)
	var topSections []string
	topSections = append(topSections, sectionHeader.Render("Queue Information"))
	topSections = append(topSections, row("Queue Name", m.state.queueMessageCreate.queueName))

	queueType := "Standard"
	if m.state.queueMessageCreate.isFifo {
		queueType = "FIFO"
	}
	topSections = append(topSections, row("Queue Type", queueType))

	topContent := lipgloss.JoinVertical(lipgloss.Left, topSections...)

	// Left panel - bottom section (instructions and buttons)
	var bottomSections []string
	bottomSections = append(bottomSections, sectionHeader.Render("Instructions"))
	instructionStyle := lipgloss.NewStyle().
		Foreground(styles.MediumGray).
		Width(leftContentWidth).
		MarginTop(1)
	bottomSections = append(bottomSections, instructionStyle.Render("Enter your message body in the text area on the right."))
	bottomSections = append(bottomSections, instructionStyle.Render("Supports JSON or plain text up to 256KB."))
	if m.state.queueMessageCreate.isFifo {
		bottomSections = append(bottomSections, instructionStyle.MarginTop(1).Render("FIFO messages will use 'default' as the message group ID."))
	}

	// Buttons
	cancelBtn := "cancel"
	submitBtn := "submit"

	switch m.state.queueMessageCreate.selected {
	case 1:
		cancelBtn = styles.ButtonSecondary.Render(cancelBtn)
		submitBtn = styles.ButtonPrimary.Render(submitBtn)
	case 2:
		cancelBtn = styles.ButtonPrimary.Render(cancelBtn)
		submitBtn = styles.ButtonSecondary.Render(submitBtn)
	default:
		cancelBtn = styles.ButtonPrimary.Render(cancelBtn)
		submitBtn = styles.ButtonPrimary.Render(submitBtn)
	}

	buttonRow := lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "    ", submitBtn)
	buttons := lipgloss.NewStyle().
		MarginTop(2).
		Width(leftContentWidth).
		Render(lipgloss.PlaceHorizontal(leftContentWidth, lipgloss.Center, buttonRow))

	bottomSections = append(bottomSections, buttons)
	bottomContent := lipgloss.JoinVertical(lipgloss.Left, bottomSections...)

	// Combine top and bottom with bottom aligned to the bottom
	leftPanelInner := lipgloss.JoinVertical(lipgloss.Left,
		topContent,
		lipgloss.PlaceVertical(contentHeight-lipgloss.Height(topContent), lipgloss.Bottom, bottomContent),
	)

	leftPanelStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2).
		Width(messageLeftPanelWidth).
		Height(contentHeight)

	leftPanel := leftPanelStyle.Render(leftPanelInner)

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

	// Right panel - message body textarea
	bodyHeaderStyle := lipgloss.NewStyle().
		Foreground(styles.TextLight).
		Bold(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(styles.BorderColor).
		Width(rightContentWidth).
		MarginBottom(1)

	rightPanelStyle := lipgloss.NewStyle().
		PaddingLeft(2).
		PaddingRight(2).
		Width(messageRightPanelWidth).
		Height(contentHeight)

	rightContent := lipgloss.JoinVertical(lipgloss.Left,
		bodyHeaderStyle.Render("Message Body"),
		m.state.queueMessageCreate.textarea.View(),
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

func (m model) QueueMessageCreateUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m.QueueDetailsGoBack(msg)

		case msg.Type == tea.KeyTab || msg.Type == tea.KeyShiftTab:
			if msg.Type == tea.KeyShiftTab {
				m.state.queueMessageCreate.selected--
				if m.state.queueMessageCreate.selected < 0 {
					m.state.queueMessageCreate.selected = 2
				}
			} else {
				m.state.queueMessageCreate.selected = (m.state.queueMessageCreate.selected + 1) % 3
			}

			if m.state.queueMessageCreate.selected == 0 {
				m.state.queueMessageCreate.textarea.Focus()
			} else {
				m.state.queueMessageCreate.textarea.Blur()
			}
			return m, nil

		case key.Matches(msg, m.keys.View):
			switch m.state.queueMessageCreate.selected {
			case 1:
				return m.QueueDetailsGoBack(msg)
			case 2:
				body := strings.TrimSpace(m.state.queueMessageCreate.textarea.Value())
				if body == "" {
					m.error = "Message body cannot be empty"
					return m, nil
				}

				m.loading = true
				m.loadingMsg = "Sending message..."

				input := kue.SendMessageInput{
					QueueUrl:    m.state.queueMessageCreate.queueUrl,
					MessageBody: body,
				}
				if m.state.queueMessageCreate.isFifo {
					input.MessageGroupId = "default"
				}
				return m, commands.SendMessage(m.context, m.client, input)
			}
		}

		if m.state.queueMessageCreate.selected == 0 {
			m.state.queueMessageCreate.textarea, cmd = m.state.queueMessageCreate.textarea.Update(msg)
			return m, cmd
		}
	}

	if m.state.queueMessageCreate.selected == 0 {
		m.state.queueMessageCreate.textarea, cmd = m.state.queueMessageCreate.textarea.Update(msg)
	}

	return m, cmd
}
