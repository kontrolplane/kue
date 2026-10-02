package tui

import (
	"fmt"
	"slices"
	"strings"

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
	fields    viewport.Model // the queue, message and attributes, which can be more than fit
	onFields  bool           // the keys scroll the fields rather than the body
}

func (m model) QueueMessageDetailsSwitchPage() (model, tea.Cmd) {
	m.error = ""
	d := &m.state.queueMessageDetails
	d.body = newPayloadText(d.message.Body)
	d.kind = payloadKind(d.message.Body)
	d.viewport = viewport.New(viewport.WithWidth(rightContentWidth), viewport.WithHeight(detailsViewportHeight()))
	d.viewport.SetContent(d.body.render(rightContentWidth))
	d.fields = viewport.New(viewport.WithWidth(leftContentWidth), viewport.WithHeight(contentHeight))
	d.fields.SetContent(d.renderFields())
	d.onFields = false
	return m.SwitchPage(queueMessageDetails), nil
}

// resizeBody wraps the body again at the current width, keeping the scroll positions.
func (d *queueMessageDetailsState) resizeBody() {
	offset := d.viewport.YOffset()
	d.viewport.SetWidth(rightContentWidth)
	d.viewport.SetHeight(detailsViewportHeight())
	d.viewport.SetContent(d.body.render(rightContentWidth))
	d.viewport.SetYOffset(offset)
	offset = d.fields.YOffset()
	d.fields.SetWidth(leftContentWidth)
	d.fields.SetHeight(contentHeight)
	d.fields.SetContent(d.renderFields())
	d.fields.SetYOffset(offset)
	if !d.fieldsOverflow() {
		d.onFields = false
	}
}

// fieldsOverflow reports whether the fields are more than the panel shows, so it scrolls.
func (d queueMessageDetailsState) fieldsOverflow() bool {
	return d.fields.TotalLineCount() > d.fields.Height()
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd
	d := &m.state.queueMessageDetails
	// The fields hold how long ago the message was sent, which changes while it is open.
	d.fields.SetContent(d.renderFields())
	scrolled := &d.viewport
	if d.onFields {
		scrolled = &d.fields
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(keyMsg, m.keys.NextField, m.keys.PrevField):
			d.onFields = !d.onFields && d.fieldsOverflow()
			return m, nil
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
			scrolled.GotoTop()
			return m, nil
		case key.Matches(keyMsg, m.keys.Bottom):
			scrolled.GotoBottom()
			return m, nil
		}
	}

	*scrolled, cmd = scrolled.Update(msg)
	return m, cmd
}

// systemAttributes are the message attributes the left panel already shows in its own rows.
var systemAttributes = []string{
	"SentTimestamp", "ApproximateFirstReceiveTimestamp", "ApproximateReceiveCount",
	"MessageGroupId", "MessageDeduplicationId", "SequenceNumber",
}

// renderFields renders the left panel: the queue, the message, its fifo details and attributes.
func (d queueMessageDetailsState) renderFields() string {
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

	title := styles.Fg(styles.ToneText).Bold(true)
	if d.onFields {
		title = styles.Fg(styles.ToneAccent).Bold(true)
	}
	meta := ""
	if d.fieldsOverflow() {
		meta = styles.Faint("tab scrolls")
		if d.onFields {
			meta = styles.Faint(fmt.Sprintf("%d%%", int(d.fields.ScrollPercent()*100)))
		}
	}
	left := []string{
		styles.SectionHeaderWith(title.Render("queue"), meta, leftContentWidth),
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
	left = append(left, attributeRows(msg.MessageAttributes)...)

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

	return strings.Join(left, "\n")
}

func (m model) QueueMessageDetailsView() string {
	d := m.state.queueMessageDetails
	fields := d.fields
	fields.SetContent(d.renderFields())

	vp := d.viewport
	meta := d.kind
	if vp.TotalLineCount() > vp.Height() {
		meta += fmt.Sprintf(" · %d%%", int(vp.ScrollPercent()*100))
	}
	title := styles.Fg(styles.ToneText).Bold(true)
	if !d.onFields && d.fieldsOverflow() {
		title = styles.Fg(styles.ToneAccent).Bold(true)
	}
	right := lipgloss.JoinVertical(lipgloss.Left,
		styles.SectionHeaderWith(title.Render("body"), styles.Faint(meta), rightContentWidth),
		"",
		vp.View(),
	)

	return splitPanels(fields.View(), right)
}

// attributeRows lists the message attributes on one line each.
func attributeRows(attrs map[string]string) []string {
	if len(attrs) == 0 {
		return []string{panelRowSpans("", styles.S("no attributes", styles.ToneFaint))}
	}
	names := make([]string, 0, len(attrs))
	for name := range attrs {
		names = append(names, name)
	}
	slices.Sort(names)

	lines := make([]string, 0, len(names))
	for _, name := range names {
		lines = append(lines, panelRow(name, attrs[name]))
	}
	return lines
}
