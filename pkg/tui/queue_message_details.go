package tui

import (
	"fmt"
	"slices"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

func detailsViewportHeight() int { return contentHeight - 2 } // Account for the header and its spacing

// queueMessageDetailsState holds the state for message details view.
type queueMessageDetailsState struct {
	message   kue.Message
	queueName string
	isFifo    bool
	body      payloadText
	kind      string // what the body is and its size, for the section header
	viewport  viewport.Model
}

func (m model) QueueMessageDetailsSwitchPage() (model, tea.Cmd) {
	m.error = ""
	d := &m.state.queueMessageDetails
	d.body = newPayloadText(d.message.Body)
	d.kind = payloadKind(d.message.Body)
	d.viewport = viewport.New(viewport.WithWidth(rightContentWidth), viewport.WithHeight(detailsViewportHeight()))
	d.viewport.SetContent(d.body.render(rightContentWidth))
	return m.SwitchPage(queueMessageDetails), nil
}

// resizeBody wraps the body again at the current width, keeping the scroll position.
func (d *queueMessageDetailsState) resizeBody() {
	offset := d.viewport.YOffset()
	d.viewport.SetWidth(rightContentWidth)
	d.viewport.SetHeight(detailsViewportHeight())
	d.viewport.SetContent(d.body.render(rightContentWidth))
	d.viewport.SetYOffset(offset)
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	d := &m.state.queueMessageDetails

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.CopyToClipboard):
			return m, commands.CopyToClipboard(d.message.Body)
		case key.Matches(keyMsg, m.keys.Delete):
			if d.message.ReceiptHandle == "" {
				return m, nil
			}
			md := &m.state.queueMessageDelete
			md.messages = []kue.Message{d.message}
			md.queueUrl = m.state.queueDetails.queue.Url
			md.queueName = d.queueName
			md.fromDetails = true
			return m.QueueMessageDeleteSwitchPage()
		case key.Matches(keyMsg, m.keys.Quit, m.keys.Back):
			return m.queueDetailsGoBack()
		case key.Matches(keyMsg, m.keys.Top):
			d.viewport.GotoTop()
			return m, nil
		case key.Matches(keyMsg, m.keys.Bottom):
			d.viewport.GotoBottom()
			return m, nil
		}
	}

	d.viewport, cmd = d.viewport.Update(msg)
	return m, cmd
}

// systemAttributes are the message attributes the left panel already shows in its own rows.
var systemAttributes = []string{
	"SentTimestamp", "ApproximateFirstReceiveTimestamp", "ApproximateReceiveCount",
	"MessageGroupId", "MessageDeduplicationId", "SequenceNumber",
}

func (m model) QueueMessageDetailsView() string {
	d := m.state.queueMessageDetails
	msg := d.message

	kind := "standard"
	if d.isFifo {
		kind = "fifo"
	}
	timestamp := func(s string) []styles.Span {
		t := parseTime(s)
		if t.IsZero() {
			return []styles.Span{styles.S("-", styles.ToneFaint)}
		}
		return []styles.Span{styles.S(formatTime(t), styles.ToneBody), styles.S("  "+formatAgo(t), styles.ToneFaint)}
	}
	receives := styles.S(orDash(msg.ReceiveCount), styles.ToneBody)
	if atoi(msg.ReceiveCount) > 1 {
		receives.Tone = styles.ToneWarning
	}

	left := []string{
		panelSection("queue", true, leftContentWidth),
		panelRowSpans("name", styles.S(d.queueName, styles.ToneText)),
		panelRow("type", kind),
		"",
		panelSection("message", true, leftContentWidth),
		panelRowSpans("id", styles.S(msg.MessageID, styles.ToneText)),
		panelRowSpans("sent", timestamp(msg.SentTimestamp)...),
		panelRowSpans("first received", timestamp(msg.FirstReceiveTime)...),
		panelRowSpans("receives", receives),
		panelRow("size", formatBytes(uint64(len(msg.Body)))),
		panelRowSpans("md5", styles.S(msg.MD5OfBody, styles.ToneMuted)),
	}

	if msg.MessageGroupID != "" || msg.MessageDeduplicationID != "" || msg.SequenceNumber != "" {
		left = append(left, "", panelSection("fifo", true, leftContentWidth))
		if msg.MessageGroupID != "" {
			left = append(left, panelRow("group id", msg.MessageGroupID))
		}
		if msg.MessageDeduplicationID != "" {
			left = append(left, panelRow("deduplication id", msg.MessageDeduplicationID))
		}
		if msg.SequenceNumber != "" {
			left = append(left, panelRow("sequence", msg.SequenceNumber))
		}
	}

	left = append(left, "", styles.SectionHeaderWith(styles.Bold("attributes"), styles.Faint(fmt.Sprint(len(msg.MessageAttributes))), leftContentWidth))
	left = append(left, attributeRows(msg.MessageAttributes, contentHeight-len(left)-2)...)

	var system []string
	for name := range msg.Attributes {
		if !slices.Contains(systemAttributes, name) {
			system = append(system, name)
		}
	}
	if len(system) > 0 {
		slices.Sort(system)
		left = append(left, "", panelSection("system attributes", true, leftContentWidth))
		for _, name := range system {
			left = append(left, panelRow(name, msg.Attributes[name]))
		}
	}

	vp := d.viewport
	meta := d.kind
	if vp.TotalLineCount() > vp.Height() {
		meta += fmt.Sprintf(" · %d%%", int(vp.ScrollPercent()*100))
	}
	right := lipgloss.JoinVertical(lipgloss.Left,
		styles.SectionHeaderWith(styles.Bold("body"), styles.Faint(meta), rightContentWidth),
		"",
		vp.View(),
	)

	return splitPanels(lipgloss.JoinVertical(lipgloss.Left, left...), right)
}

// attributeRows lists the message attributes on one line each, in at most rows lines, with the
// ones past them counted.
func attributeRows(attrs map[string]string, rows int) []string {
	if len(attrs) == 0 {
		return []string{panelRowSpans("", styles.S("no attributes", styles.ToneFaint))}
	}
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	slices.Sort(names)

	shown := names
	if rows = max(rows, 1); len(names) > rows {
		shown = names[:rows-1]
	}
	lines := make([]string, 0, len(shown)+1)
	for _, name := range shown {
		lines = append(lines, panelRow(name, attrs[name]))
	}
	if more := len(names) - len(shown); more > 0 {
		lines = append(lines, panelRowSpans("", styles.S(fmt.Sprintf("+%d more attributes", more), styles.ToneFaint)))
	}
	return lines
}
