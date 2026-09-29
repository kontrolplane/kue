package tui

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// spread places left and right on one line of the given width.
func spread(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

const (
	headerIndent = 1
	headerGap    = 2 // the least room between the fact columns
)

// headerFact is one label and value in the header grid.
type headerFact struct {
	label string
	value cell
}

// factsWidth is the width facts need to show in full.
func factsWidth(facts []headerFact) int {
	labelWidth, valueWidth := 0, 0
	for _, f := range facts {
		labelWidth = max(labelWidth, len(f.label))
		w := 0
		for _, s := range f.value {
			w += ansi.StringWidth(s.Text)
		}
		valueWidth = max(valueWidth, w)
	}
	return labelWidth + 2 + valueWidth
}

// factColumn renders facts as rows with their labels aligned, filling width.
func factColumn(width int, facts ...headerFact) string {
	labelWidth := factsLabelWidth(facts)
	lines := make([]string, len(facts))
	for i, f := range facts {
		label := styles.Faint(f.label + strings.Repeat(" ", labelWidth-len(f.label)+2))
		lines[i] = label + renderCell(f.value, column{width: max(1, width-labelWidth-2)}, nil)
	}
	return strings.Join(lines, "\n")
}

// renderHeader draws the brand and refresh state on top, with a grid of connection, account and
// queue facts below it.
func (m model) renderHeader() string {
	width := frameWidth - headerIndent
	indent := strings.Repeat(" ", headerIndent)

	brand := styles.Render(
		styles.S(m.projectName+"/", styles.ToneFaint),
		styles.B(m.programName, styles.ToneText),
	)
	status := []styles.Span{styles.S("refresh every "+compactDuration(commands.RefreshInterval), styles.ToneFaint)}
	if m.paused {
		status = []styles.Span{styles.S("⏸ paused", styles.ToneWarning)}
	}
	top := spread(brand, styles.Render(status...), width)

	facts := factsGrid(width, m.connectionFacts(), m.accountFacts(), m.queueFacts(formatCount), m.messageFacts(formatCount))
	if facts == "" {
		facts = factsGrid(width, m.connectionFacts(), m.accountFacts(), m.queueFacts(compactCount), m.messageFacts(compactCount))
	}
	if facts == "" {
		facts = factsGrid(-1, m.connectionFacts(), m.queueFacts(compactCount), m.messageFacts(compactCount))
	}

	return indent + top + "\n" + lipgloss.NewStyle().MarginLeft(headerIndent).Render(facts)
}

// factsGrid lays the fact columns out across width, or returns "" when they do not fit. Spare
// room goes between the columns so the grid spans the header like the line above it. Without
// any, the columns before the last two give up width, since names and urls read fine cut short
// while the numbers do not, but never below a floor that keeps them readable. A nil column is
// left out, and a negative width lays the columns out whether they fit or not.
func factsGrid(width int, columns ...[]headerFact) string {
	columns = slices.DeleteFunc(columns, func(f []headerFact) bool { return f == nil })
	natural := make([]int, len(columns))
	floor := make([]int, len(columns))
	total := 0
	for i, facts := range columns {
		natural[i] = factsWidth(facts)
		floor[i] = natural[i]
		total += natural[i]
	}
	// The last two columns hold numbers, the ones before them names that read fine cut short.
	flexible := len(columns) - 2
	for i := range flexible {
		floor[i] = min(natural[i], factsLabelWidth(columns[i])+2+12)
	}

	gaps := len(columns) - 1
	spare := width - total
	if short := gaps*headerGap - spare; short > 0 && width >= 0 {
		room := 0
		for i := range flexible {
			room += natural[i] - floor[i]
		}
		if short > room {
			return ""
		}
		left := short
		for i := flexible - 1; i >= 0 && left > 0; i-- {
			give := min(left, natural[i]-floor[i])
			natural[i] -= give
			left -= give
		}
		spare = gaps * headerGap
	}
	spare = max(spare, gaps)

	rendered := make([]string, 0, 2*len(columns)-1)
	for i, facts := range columns {
		if i > 0 {
			gap := spare / gaps
			if i <= spare%gaps {
				gap++
			}
			rendered = append(rendered, strings.Repeat(" ", gap))
		}
		rendered = append(rendered, factColumn(natural[i], facts...))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

func factsLabelWidth(facts []headerFact) int {
	w := 0
	for _, f := range facts {
		w = max(w, len(f.label))
	}
	return w
}

// pendingValue stands in for facts that are not loaded yet.
var pendingValue = text("…", styles.ToneFaint)

func (m model) connectionFacts() []headerFact {
	return []headerFact{
		{"profile", text(orDash(m.awsInfo.Profile), styles.ToneBody)},
		{"region", text(orDash(m.awsInfo.Region), styles.ToneBody)},
	}
}

func (m model) accountFacts() []headerFact {
	endpoint := text("aws", styles.ToneBody)
	if m.awsInfo.Endpoint != "" {
		endpoint = text(m.awsInfo.Endpoint, styles.ToneBody)
	}
	account := pendingValue
	if m.state.queueOverview.loaded {
		account = text("-", styles.ToneFaint)
		for _, q := range m.state.queueOverview.queues {
			if id := accountID(q.Arn); id != "" {
				account = text(id, styles.ToneBody)
				break
			}
		}
	}
	return []headerFact{{"account", account}, {"endpoint", endpoint}}
}

// accountID reads the account from a queue arn, arn:aws:sqs:region:account:name.
func accountID(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) < 6 {
		return ""
	}
	return parts[4]
}

// queueFacts counts the queues and the dead-letter queues among them, written out by format.
func (m model) queueFacts(format func(uint64) string) []headerFact {
	o := m.state.queueOverview
	if !o.loaded {
		return []headerFact{{"queues", pendingValue}, {"dead-letter", pendingValue}}
	}
	dlqs := 0
	for _, q := range o.queues {
		if isDeadLetter(o.queues, q) {
			dlqs++
		}
	}
	n := func(v int) cell {
		if v == 0 {
			return text("0", styles.ToneFaint)
		}
		return text(format(uint64(v)), styles.ToneBody)
	}
	return []headerFact{{"queues", n(len(o.queues))}, {"dead-letter", n(dlqs)}}
}

// messageFacts sums the messages available and in flight across the queues.
func (m model) messageFacts(format func(uint64) string) []headerFact {
	o := m.state.queueOverview
	if !o.loaded {
		return []headerFact{{"available", pendingValue}, {"in flight", pendingValue}}
	}
	var available, inFlight uint64
	for _, q := range o.queues {
		available += atoi(q.ApproximateNumberOfMessages)
		inFlight += atoi(q.ApproximateNumberOfMessagesNotVisible)
	}
	n := func(v uint64, tone styles.Tone) cell {
		if v == 0 {
			return text("0", styles.ToneFaint)
		}
		return text(format(v), tone)
	}
	return []headerFact{{"available", n(available, styles.ToneBody)}, {"in flight", n(inFlight, styles.ToneInfo)}}
}

// isDeadLetter reports whether another queue sends its failed messages to q.
func isDeadLetter(queues []kue.Queue, q kue.Queue) bool {
	return q.Arn != "" && slices.ContainsFunc(queues, func(o kue.Queue) bool { return o.DeadLetterTargetARN == q.Arn })
}

// maxCrumbWidth bounds the crumbs between the first and the last, so a long queue name leaves
// room for the page it leads to.
const maxCrumbWidth = 48

// breadcrumb names where the current page sits, e.g. queues › orders › messages › 1f3c….
func (m model) breadcrumb() []styles.Span {
	queue := m.state.queueDetails.queue.Name
	var trail []string
	switch m.page {
	case queueOverview:
		trail = []string{"queues"}
	case queueCreate:
		trail = []string{"queues", "new queue"}
	case queueDelete:
		trail = []string{"queues", "delete"}
	case queuePurge:
		trail = []string{"queues", m.state.queuePurge.queue.Name, "purge"}
	case queueRedrive:
		trail = []string{"queues", m.state.queueRedrive.queue.Name, "redrive"}
	case queueDetails:
		trail = []string{"queues", queue}
	case queueMessageDetails:
		trail = []string{"queues", queue, "messages", truncate(m.state.queueMessageDetails.message.MessageID, 13)}
	case queueMessageCreate:
		trail = []string{"queues", queue, "send"}
	case queueMessageDelete:
		trail = []string{"queues", queue, "messages", "delete"}
	}

	var spans []styles.Span
	for i, t := range trail {
		if i > 0 {
			spans = append(spans, styles.S(" › ", styles.ToneFaint))
		}
		if i > 0 && i < len(trail)-1 {
			t = truncate(t, maxCrumbWidth)
		}
		if i == len(trail)-1 {
			spans = append(spans, styles.B(t, styles.ToneText))
		} else {
			spans = append(spans, styles.S(t, styles.ToneMuted))
		}
	}
	if m.page == queueOverview && len(m.state.queueOverview.queues) > 0 {
		spans = append(spans, styles.S(" "+formatCount(uint64(len(m.state.queueOverview.queues))), styles.ToneFaint))
	}
	return spans
}

// clip cuts s to height lines, ending lines wider than width in an ellipsis.
func clip(s string, width, height int) string {
	lines := strings.Split(s, "\n")
	if len(lines) > height {
		lines = lines[:max(height, 0)]
	}
	for i, line := range lines {
		if ansi.StringWidth(line) > width {
			lines[i] = ansi.Truncate(line, width, "…")
		}
	}
	return strings.Join(lines, "\n")
}

// frame draws the rounded border around the content area, with a title set into the top edge
// and meta text into the bottom one.
func frame(title, meta, foot, body string) string {
	border := lipgloss.NewStyle().Foreground(styles.P.Rule)
	edge := func(n int) string {
		if n < 1 {
			n = 1
		}
		return border.Render(strings.Repeat("─", n))
	}

	metaWidth := 0
	if meta != "" {
		meta = " " + ansi.Truncate(meta, frameWidth/3, "…") + " "
		metaWidth = lipgloss.Width(meta)
	}
	title = ansi.Truncate(title, frameWidth-7-metaWidth, "…")
	top := border.Render("╭─ ") + title + " " +
		edge(frameWidth-5-lipgloss.Width(title)-1-metaWidth) + meta + border.Render("─╮")

	footWidth := 0
	if foot != "" {
		foot = " " + ansi.Truncate(foot, frameWidth/2, "…") + " "
		footWidth = lipgloss.Width(foot)
	}
	bottom := border.Render("╰") + edge(frameWidth-3-footWidth) + foot + border.Render("─╯")

	placed := lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Top, clip(body, contentWidth, contentHeight))
	lines := append([]string{""}, strings.Split(placed, "\n")...)
	lines = append(lines, "")

	var b strings.Builder
	b.WriteString(top)
	for _, line := range lines {
		b.WriteString("\n")
		b.WriteString(border.Render("│"))
		b.WriteString(line + strings.Repeat(" ", max(0, contentWidth-lipgloss.Width(line))))
		b.WriteString(border.Render("│"))
	}
	b.WriteString("\n")
	b.WriteString(bottom)
	return b.String()
}
