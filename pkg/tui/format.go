package tui

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

const timeFormat = "2006-01-02 15:04:05"

// parseTime reads the timestamps the kue package formats: queue timestamps in local time, message
// timestamps as RFC 3339.
func parseTime(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.ParseInLocation(timeFormat, s, time.Local); err == nil {
		return t
	}
	return time.Time{}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format(timeFormat)
}

// formatAgo renders a timestamp relative to now, e.g. "3m ago".
func formatAgo(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := max(time.Since(t), 0) // a clock running ahead reads as 0s ago
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatCount(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if len(s) <= 3 {
		return s
	}
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < len(s); i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// compactCount renders a count exactly up to 9,999,999 and shortened past it, e.g. 12.3M or 4.5B,
// so counts fit the columns of a table. It rounds down, so a count never reads as more than it is.
func compactCount(n uint64) string {
	if n <= 9_999_999 {
		return formatCount(n)
	}
	v, unit := float64(n)/1e6, "M"
	for _, u := range []string{"B", "T", "Q"} {
		if v < 1000 {
			break
		}
		v, unit = v/1000, u
	}
	if v < 100 {
		return strconv.FormatFloat(math.Floor(v*10)/10, 'f', 1, 64) + unit
	}
	return strconv.FormatFloat(math.Floor(v), 'f', 0, 64) + unit
}

// atoi reads a count SQS reports as a string, 0 when it is missing.
func atoi(s string) uint64 {
	n, _ := strconv.ParseUint(s, 10, 64)
	return n
}

// formatSeconds renders an attribute given in seconds as a compact duration, e.g. 4d or 30s.
func formatSeconds(s string) string {
	n, err := strconv.Atoi(s)
	if err != nil {
		return orDash(s)
	}
	return compactDuration(time.Duration(n) * time.Second)
}

// compactDuration renders a duration without its zero units, e.g. 1h rather than 1h0m0s, and
// whole days as days.
func compactDuration(d time.Duration) string {
	if d > 0 && d%(24*time.Hour) == 0 {
		return fmt.Sprintf("%dd", d/(24*time.Hour))
	}
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}

// preview renders message data on a single line for use in tables. Only the prefix that can be
// shown is scanned, so its cost does not grow with the payload.
func preview(data string, width int) string {
	var b strings.Builder
	space := false
	for i := 0; i < len(data) && b.Len() <= width*4; {
		r, n := utf8.DecodeRuneInString(data[i:])
		i += n
		if unicode.IsSpace(r) {
			space = b.Len() > 0
			continue
		}
		if space {
			b.WriteByte(' ')
			space = false
		}
		b.WriteRune(r)
	}
	return truncate(b.String(), width)
}

func truncate(s string, width int) string {
	return ansi.Truncate(s, width, "…")
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func plural(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%s %s", formatCount(uint64(n)), noun)
	}
	return fmt.Sprintf("%s %ss", formatCount(uint64(n)), noun)
}

// formatPayload pretty prints JSON payloads and returns anything else as is, reporting whether
// the payload is JSON.
func formatPayload(body string) (string, bool) {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		return body, false
	}
	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return body, false
	}
	return string(pretty), true
}

func payloadKind(body string) string {
	switch {
	case body == "":
		return "empty"
	case json.Valid([]byte(body)):
		return "json · " + formatBytes(uint64(len(body)))
	default:
		return "text · " + formatBytes(uint64(len(body)))
	}
}

// payloadText is a message body as the details view shows it: pretty printed when it is JSON
// and cleaned.
type payloadText struct {
	text  string
	json  bool
	empty bool
}

func newPayloadText(body string) payloadText {
	if body == "" {
		return payloadText{empty: true}
	}
	text, js := formatPayload(body)
	return payloadText{text: styles.CleanBlock(text), json: js}
}

// render wraps the payload at width, or not at all when width is 0, and colours it.
func (p payloadText) render(width int) string {
	if p.empty {
		return styles.Faint("empty body")
	}
	text := p.text
	if width > 0 {
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			lines[i] = strings.Join(wrapLine(line, width), "\n")
		}
		text = strings.Join(lines, "\n")
	}
	if p.json {
		return highlightJSON(text)
	}
	var b strings.Builder
	paintLines(&b, styles.ToneBody, text)
	return b.String()
}

// wrapLine breaks a line into pieces of at most width columns, after a space or a comma where one
// is in the second half of a piece, so words and values stay whole, and anywhere otherwise. The
// pieces hold every byte of the line, in order, so the colours of JSON stay on their tokens.
func wrapLine(s string, width int) []string {
	if textWidth(s) <= width {
		return []string{s}
	}
	var out []string
	for textWidth(s) > width {
		head := truncateText(s, width)
		at := len(head)
		if i := strings.LastIndexAny(head, " ,"); i >= len(head)/2 {
			at = i + 1
		}
		if at == 0 {
			at = len(s) // a rune wider than the line
		}
		out = append(out, s[:at])
		s = s[at:]
	}
	if s != "" {
		out = append(out, s)
	}
	return out
}

// paintLines writes s in tone, styling each line on its own so a wrapped token keeps its colour
// on every line without being padded into a block.
func paintLines(b *strings.Builder, tone styles.Tone, s string) {
	prefix := sgr(styles.P.Color(tone), false, nil)
	for i, line := range strings.Split(s, "\n") {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(prefix)
		b.WriteString(line)
		b.WriteString(ansi.ResetStyle)
	}
}

// highlightJSON colours pretty printed JSON: keys carry the text, strings the accent, numbers and
// literals the warm ink, and punctuation recedes.
func highlightJSON(s string) string {
	var b strings.Builder
	paint := func(tone styles.Tone, t string) { paintLines(&b, tone, t) }

	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c == '"':
			j := i + 1
			for j < len(s) && s[j] != '"' {
				if s[j] == '\\' {
					j++
				}
				j++
			}
			j = min(j+1, len(s))
			k := j
			for k < len(s) && s[k] == ' ' {
				k++
			}
			if k < len(s) && s[k] == ':' {
				paint(styles.ToneText, s[i:j])
			} else {
				paint(styles.ToneAccent, s[i:j])
			}
			i = j
		case c == '-' || (c >= '0' && c <= '9'):
			j := i + 1
			for j < len(s) && strings.IndexByte("0123456789.eE+-", s[j]) >= 0 {
				j++
			}
			paint(styles.ToneWarm, s[i:j])
			i = j
		case c == 't' || c == 'f' || c == 'n':
			j := i
			for j < len(s) && s[j] >= 'a' && s[j] <= 'z' {
				j++
			}
			paint(styles.ToneWarm, s[i:j])
			i = max(j, i+1)
		case c == ' ' || c == '\n':
			b.WriteByte(c)
			i++
		default:
			paint(styles.ToneFaint, string(c))
			i++
		}
	}
	return b.String()
}

func initFilterInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = ""
	ti.CharLimit = 100
	ti.SetWidth(50)
	ti.SetStyles(styles.TextInput())
	return ti
}

// setRows replaces the rows of a table and returns the cursor clamped to the new row count.
func setRows(t *dataTable, rows []tableRow, cursor int) int {
	t.SetRows(rows)
	t.SetCursor(cursor)
	return t.Cursor()
}

func verticalDivider(height int) string {
	lines := make([]string, height)
	for i := range lines {
		lines[i] = "│"
	}
	return lipgloss.NewStyle().Foreground(styles.P.Rule).Render(strings.Join(lines, "\n"))
}

// tableCount renders a count for a table cell, receding when it is zero and shortened past
// 9,999,999 so it fits its column.
func tableCount(n uint64, tone styles.Tone) cell {
	if n == 0 {
		return text("0", styles.ToneFaint)
	}
	return text(compactCount(n), tone)
}

// lastActivity renders how long ago something happened, lit up while it is still fresh. A fresh
// time is padded to the width of "59s ago" so the dots line up in a right-aligned column.
func lastActivity(t time.Time) cell {
	switch {
	case t.IsZero():
		return text("-", styles.ToneFaint)
	case time.Since(t) < time.Minute:
		return cell{styles.S("● ", styles.ToneSuccess), styles.S(fmt.Sprintf("%7s", formatAgo(t)), styles.ToneBody)}
	default:
		return text(formatAgo(t), styles.ToneMuted)
	}
}

func sep() styles.Span { return styles.S(" · ", styles.ToneFaint) }
