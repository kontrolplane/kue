package tui

import (
	"image/color"
	"slices"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

type column struct {
	title string
	width int
	right bool
	grow  int // share of the width left over once every column has its own
	min   int // narrowest a growing column gets, max(len(title), minColumnWidth) when unset
	drop  int // when the table cannot fit, columns with the highest drop are hidden first; 0 stays
	hide  bool
}

// minimum is the narrowest the column gets.
func (c column) minimum() int {
	if c.grow == 0 {
		return c.width
	}
	return max(c.min, len(c.title), minColumnWidth)
}

type cell []styles.Span

type tableRow []cell

// text builds a cell from a single span.
func text(s string, tone styles.Tone) cell { return cell{styles.S(s, tone)} }

const (
	tableGap       = 2
	tableGutter    = 2 // room for the cursor marker
	minColumnWidth = 8
)

// dataTable is a scrolling table whose cells keep their own tones. The bubbles table paints a
// selected row with a single style, which a coloured cell would cut through; this one paints every
// span onto the row's ground.
type dataTable struct {
	base    []column // columns at their own widths, before growing into the table width
	columns []column
	rows    []tableRow
	cursor  int
	offset  int
	height  int // body rows, excluding the header
	focused bool
	empty   string
}

func newDataTable(columns []column, width, height int) dataTable {
	t := dataTable{base: columns, focused: true}
	t.setSize(width, height)
	return t
}

func (t dataTable) width() int {
	return columnsWidth(t.columns)
}

// columnsWidth is the width of the visible columns with the gutter and the gaps between them.
func columnsWidth(columns []column) int {
	return sumWidths(columns, func(c column) int { return c.width })
}

// minColumnsWidth is columnsWidth with every growing column at its minimum.
func minColumnsWidth(columns []column) int {
	return sumWidths(columns, column.minimum)
}

func sumWidths(columns []column, width func(column) int) int {
	w := tableGutter
	first := true
	for _, c := range columns {
		if c.hide {
			continue
		}
		if !first {
			w += tableGap
		}
		first = false
		w += width(c)
	}
	return w
}

// setSize fits the columns to width and shows height body rows. While the columns do not fit at
// their minimum, the ones with the highest drop are hidden. What is left over or missing is shared
// by the growing columns.
func (t *dataTable) setSize(width, height int) {
	t.height = max(height, minTableHeight)
	t.columns = slices.Clone(t.base)
	for minColumnsWidth(t.columns) > width {
		worst := -1
		for i, c := range t.columns {
			if !c.hide && c.drop > 0 && (worst < 0 || c.drop > t.columns[worst].drop) {
				worst = i
			}
		}
		if worst < 0 {
			break
		}
		t.columns[worst].hide = true
	}

	shares := 0
	for _, c := range t.columns {
		if !c.hide {
			shares += c.grow
		}
	}
	if extra := width - columnsWidth(t.columns); extra != 0 && shares > 0 {
		given := 0
		last := -1
		for i, c := range t.columns {
			if c.grow > 0 && !c.hide {
				n := extra * c.grow / shares
				t.columns[i].width += n
				given += n
				last = i
			}
		}
		t.columns[last].width += extra - given
		for i, c := range t.columns {
			t.columns[i].width = max(c.width, c.minimum())
		}
	}
	t.SetCursor(t.cursor)
}

func (t dataTable) Rows() []tableRow { return t.rows }
func (t dataTable) Cursor() int      { return t.cursor }
func (t *dataTable) Focus()          { t.focused = true }
func (t *dataTable) Blur()           { t.focused = false }

func (t *dataTable) SetRows(rows []tableRow) {
	t.rows = rows
	t.SetCursor(t.cursor)
}

func (t *dataTable) SetCursor(i int) {
	t.cursor = max(0, min(i, len(t.rows)-1))
	if t.cursor < t.offset {
		t.offset = t.cursor
	}
	if t.cursor >= t.offset+t.height {
		t.offset = t.cursor - t.height + 1
	}
	t.offset = max(0, min(t.offset, len(t.rows)-t.height))
}

func (t dataTable) Update(msg tea.KeyPressMsg) dataTable {
	switch msg.String() {
	case "up", "k":
		t.SetCursor(t.cursor - 1)
	case "down", "j":
		t.SetCursor(t.cursor + 1)
	case "pgup":
		t.SetCursor(t.cursor - t.height)
	case "pgdown":
		t.SetCursor(t.cursor + t.height)
	case "home", "g":
		t.SetCursor(0)
	case "end", "G":
		t.SetCursor(len(t.rows) - 1)
	}
	return t
}

// position describes the scroll position, e.g. "12 of 240".
func (t dataTable) position() string {
	if len(t.rows) == 0 {
		return ""
	}
	return formatCount(uint64(t.cursor+1)) + " of " + formatCount(uint64(len(t.rows)))
}

func (t dataTable) View() string {
	var b strings.Builder

	header := strings.Repeat(" ", tableGutter)
	first := true
	for _, c := range t.columns {
		if c.hide {
			continue
		}
		if !first {
			header += strings.Repeat(" ", tableGap)
		}
		first = false
		header += renderCell(text(c.title, styles.ToneFaint), c, nil)
	}
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(styles.P.Rule).Render(strings.Repeat("─", t.width())))

	for i := 0; i < t.height; i++ {
		b.WriteString("\n")
		r := t.offset + i
		switch {
		case r < len(t.rows):
			b.WriteString(t.renderRow(r))
		case i == 1 && len(t.rows) == 0 && t.empty != "":
			b.WriteString(strings.Repeat(" ", tableGutter) + styles.Muted(truncate(styles.Clean(t.empty), t.width()-tableGutter)))
		}
	}
	return b.String()
}

func (t dataTable) renderRow(r int) string {
	selected := r == t.cursor
	var bg color.Color
	marker := strings.Repeat(" ", tableGutter)
	if selected {
		ground := styles.P.Surface
		markerTone := styles.ToneMuted
		if t.focused {
			ground = styles.P.Selection
			markerTone = styles.ToneAccent
		}
		bg = ground
		marker = styles.S("▌", markerTone).On(ground) + lipgloss.NewStyle().Background(ground).Render(" ")
	}

	var b strings.Builder
	b.WriteString(marker)
	first := true
	for i, c := range t.columns {
		if c.hide {
			continue
		}
		if !first {
			b.WriteString(pad(tableGap, bg))
		}
		first = false
		var content cell
		if i < len(t.rows[r]) {
			content = t.rows[r][i]
		}
		if selected {
			content = lift(content)
		}
		b.WriteString(renderCell(content, c, bg))
	}
	return b.String()
}

// lift raises faint spans on a selected row, where the ground has less contrast to spare.
func lift(c cell) cell {
	out := make(cell, len(c))
	for i, s := range c {
		switch s.Tone {
		case styles.ToneFaint:
			s.Tone = styles.ToneMuted
		case styles.ToneRule:
			s.Tone = styles.ToneFaint
		}
		out[i] = s
	}
	return out
}

func pad(n int, bg color.Color) string {
	if n <= 0 {
		return ""
	}
	if bg == nil {
		return strings.Repeat(" ", n)
	}
	return sgr(nil, false, bg) + strings.Repeat(" ", n) + ansi.ResetStyle
}

// renderCell truncates the spans to the column width and aligns them. Content that does not fit
// always ends in an ellipsis, also when the cut falls between two spans.
func renderCell(c cell, col column, bg color.Color) string {
	total := 0
	for _, s := range c {
		total += textWidth(s.Text)
	}
	budget := col.width
	cut := total > col.width
	if cut {
		budget = col.width - 1
	}

	var b strings.Builder
	used := 0
	var last styles.Span
	for _, s := range c {
		if used >= budget {
			break
		}
		txt := s.Text
		if w := textWidth(txt); used+w > budget {
			txt = truncateText(txt, budget-used)
			used += textWidth(txt)
		} else {
			used += w
		}
		writeSpan(&b, s.Tone, s.Bold, bg, txt)
		last = s
	}
	if cut && col.width > 0 {
		if last.Text == "" && len(c) > 0 {
			last = c[0]
		}
		writeSpan(&b, last.Tone, last.Bold, bg, "…")
		used++
	}
	fill := pad(col.width-used, bg)
	if col.right {
		return fill + b.String()
	}
	return b.String() + fill
}

// writeSpan writes txt in the tone, as Span.Style().Background(bg).Render would, with the style
// sequence taken from a cache. Tables render every visible cell on every frame.
func writeSpan(b *strings.Builder, tone styles.Tone, bold bool, bg color.Color, txt string) {
	b.WriteString(sgr(styles.P.Color(tone), bold, bg))
	b.WriteString(txt)
	b.WriteString(ansi.ResetStyle)
}

type sgrKey struct {
	fg, bg color.Color
	bold   bool
}

var (
	sgrMu    sync.Mutex
	sgrCache = map[sgrKey]string{}
)

// sgr returns the style sequence lipgloss would open text in the given colours with.
func sgr(fg color.Color, bold bool, bg color.Color) string {
	k := sgrKey{fg: fg, bg: bg, bold: bold}
	sgrMu.Lock()
	defer sgrMu.Unlock()
	if p, ok := sgrCache[k]; ok {
		return p
	}
	var s ansi.Style
	if bold {
		s = s.Bold()
	}
	if fg != nil {
		s = s.ForegroundColor(fg)
	}
	if bg != nil {
		s = s.BackgroundColor(bg)
	}
	p := s.String()
	sgrCache[k] = p
	return p
}
