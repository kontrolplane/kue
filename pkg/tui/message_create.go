package tui

import (
	"encoding/json"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

type messageCreateFocus int

const (
	messageCreateFocusBody messageCreateFocus = iota
	messageCreateFocusCancel
	messageCreateFocusSubmit
	messageCreateFocusCount
)

// maxMessageSize is the largest body SQS accepts.
const maxMessageSize = 256 << 10

func messageBodyHeight() int { return contentHeight - 2 }

// queueMessageCreateState holds the state for message creation.
type queueMessageCreateState struct {
	queueName string
	queueUrl  string
	isFifo    bool
	textarea  textarea.Model
	focus     messageCreateFocus
}

func (m model) QueueMessageCreateSwitchPage() (model, tea.Cmd) {
	m.error = ""
	c := &m.state.queueMessageCreate

	ta := textarea.New()
	ta.Placeholder = "message body, json or plain text…"
	ta.SetWidth(rightContentWidth)
	ta.SetHeight(messageBodyHeight())
	ta.CharLimit = maxMessageSize
	ta.Prompt = "│ "
	ta.SetStyles(styles.TextArea())
	c.textarea = ta

	m.armed = false
	m, cmd := m.focusMessageCreate(messageCreateFocusBody)
	return m.SwitchPage(queueMessageCreate), cmd
}

func (m model) focusMessageCreate(f messageCreateFocus) (model, tea.Cmd) {
	c := &m.state.queueMessageCreate
	c.focus = f
	if f == messageCreateFocusBody {
		return m, c.textarea.Focus()
	}
	c.textarea.Blur()
	return m, nil
}

func (m model) submitMessageCreate() (model, tea.Cmd) {
	c := m.state.queueMessageCreate
	body := c.textarea.Value()
	if strings.TrimSpace(body) == "" {
		m.error = "the message body cannot be empty."
		return m, nil
	}

	input := kue.SendMessageInput{
		QueueUrl:    c.queueUrl,
		MessageBody: body,
	}
	if c.isFifo {
		input.MessageGroupId = "default"
	}
	m.busy, m.loading, m.loadingMsg = true, true, "sending message…"
	return m, commands.SendMessage(m.context, m.client, c.queueName, input)
}

func (m model) QueueMessageCreateUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	c := &m.state.queueMessageCreate

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.Back):
			if strings.TrimSpace(c.textarea.Value()) != "" && !m.armed {
				m.armed = true
				return m.setStatus(discardPrompt, styles.ToneWarning)
			}
			return m.queueDetailsGoBack()
		case key.Matches(keyMsg, m.keys.Submit):
			return m.submitMessageCreate()
		case key.Matches(keyMsg, m.keys.NextField):
			return m.focusMessageCreate((c.focus + 1) % messageCreateFocusCount)
		case key.Matches(keyMsg, m.keys.PrevField):
			return m.focusMessageCreate((c.focus + messageCreateFocusCount - 1) % messageCreateFocusCount)
		}

		if c.focus != messageCreateFocusBody {
			switch {
			case key.Matches(keyMsg, m.keys.Left, m.keys.Right):
				if c.focus == messageCreateFocusCancel {
					c.focus = messageCreateFocusSubmit
				} else {
					c.focus = messageCreateFocusCancel
				}
			case key.Matches(keyMsg, m.keys.View):
				if c.focus == messageCreateFocusCancel {
					return m.queueDetailsGoBack()
				}
				return m.submitMessageCreate()
			}
			return m, nil
		}
	}

	if c.focus == messageCreateFocusBody {
		c.textarea, cmd = c.textarea.Update(msg)
	}
	return m, cmd
}

func (m model) QueueMessageCreateView() string {
	c := m.state.queueMessageCreate

	kind := "standard"
	if c.isFifo {
		kind = "fifo"
	}
	notes := []string{"the body is sent as is, json or plain text up to 256 KiB."}
	if c.isFifo {
		notes = append(notes, "fifo messages are sent with the message group id default.")
	}

	top := []string{
		panelSection("queue", true, leftContentWidth),
		panelRowSpans("name", styles.S(c.queueName, styles.ToneText)),
		panelRow("type", kind),
		"",
	}
	for _, note := range notes {
		top = append(top, styles.Faint(wrapLines(note, leftContentWidth, 2)))
	}
	topView := lipgloss.JoinVertical(lipgloss.Left, top...)

	buttons := lipgloss.JoinHorizontal(lipgloss.Center,
		styles.Button("cancel", c.focus == messageCreateFocusCancel, styles.ToneText),
		"    ",
		styles.Button("send", c.focus == messageCreateFocusSubmit, styles.ToneText),
	)
	bottom := lipgloss.PlaceHorizontal(leftContentWidth, lipgloss.Center, buttons)

	left := lipgloss.JoinVertical(lipgloss.Left,
		topView,
		lipgloss.PlaceVertical(contentHeight-lipgloss.Height(topView), lipgloss.Bottom, bottom),
	)

	bodyTitle := styles.Fg(styles.ToneText).Bold(true)
	if c.focus == messageCreateFocusBody {
		bodyTitle = styles.Fg(styles.ToneAccent).Bold(true)
	}
	right := lipgloss.JoinVertical(lipgloss.Left,
		styles.SectionHeaderWith(bodyTitle.Render("body"), bodyMeta(c.textarea.Value()), rightContentWidth),
		"",
		c.textarea.View(),
	)

	return splitPanels(left, right)
}

// bodyMeta sizes the body being written and, when it looks like JSON, says whether it parses.
func bodyMeta(body string) string {
	meta := styles.Faint(formatBytes(uint64(len(body))))
	if trimmed := strings.TrimSpace(body); strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if json.Valid([]byte(trimmed)) {
			return meta + styles.Faint(" · ") + styles.Render(styles.S("json ✓", styles.ToneSuccess))
		}
		return meta + styles.Faint(" · ") + styles.Render(styles.S("invalid json", styles.ToneWarning))
	}
	return meta
}
