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

const messageTextareaWidth = 71 // Matches section width minus borders

func (m model) QueueMessageCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""

	ta := textarea.New()
	ta.Placeholder = "Enter message body (JSON or plain text)..."
	ta.Focus()
	ta.SetWidth(messageTextareaWidth)
	ta.SetHeight(8)
	ta.CharLimit = 262144

	m.state.queueMessageCreate.textarea = ta
	m.state.queueMessageCreate.selected = 0
	return m.SwitchPage(queueMessageCreate), nil
}

func (m model) QueueMessageCreateView() string {
	const (
		labelWidth   = 22
		valueWidth   = 50
		sectionWidth = labelWidth + valueWidth + 3
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
		Width(sectionWidth).
		MarginTop(1)

	rowStyle := lipgloss.NewStyle().PaddingLeft(2)

	row := func(label, value string) string {
		return rowStyle.Render(
			lipgloss.JoinHorizontal(lipgloss.Top,
				labelStyle.Render(label),
				valueStyle.Render(value),
			),
		)
	}

	var sections []string
	sections = append(sections, sectionHeader.Render("Queue Information"))
	sections = append(sections, row("Queue Name", m.state.queueMessageCreate.queueName))

	queueType := "Standard"
	if m.state.queueMessageCreate.isFifo {
		queueType = "FIFO"
	}
	sections = append(sections, row("Queue Type", queueType))
	sections = append(sections, sectionHeader.Render("Message Body"))

	textareaStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.BorderColor).
		Padding(0, 1).
		MarginTop(1)

	if m.state.queueMessageCreate.selected == 0 {
		textareaStyle = textareaStyle.BorderForeground(styles.AccentColor)
	}

	sections = append(sections, textareaStyle.Render(m.state.queueMessageCreate.textarea.View()))

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

	buttons := lipgloss.NewStyle().
		MarginTop(2).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, cancelBtn, "    ", submitBtn))

	sections = append(sections, buttons)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
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
