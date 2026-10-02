package tui

import (
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// The view is drawn again after every message: every key, refresh and clock tick. Measuring text by
// its grapheme clusters is most of what a frame costs, while names, ids and bodies are mostly ASCII
// and the frame around them is drawn with a few symbols, all one column wide. These take that
// shortcut and leave anything else to ansi.

// narrowRanges are runs of runes that are one column wide on their own and do not join with the
// runes around them: Latin, punctuation, arrows, box drawing, blocks and shapes. A rune outside
// them, like a combining mark or an emoji, sends the whole string to ansi.
var narrowRanges = [][2]rune{
	{0x00A1, 0x00AC}, {0x00AE, 0x02FF}, // Latin-1 and extended Latin, without the soft hyphen
	{0x2010, 0x2027}, {0x2030, 0x205E}, // dashes, quotes, bullets and ellipses
	{0x2190, 0x21FF}, // arrows
	{0x2500, 0x25FC}, // box drawing, blocks and geometric shapes
	{0x2713, 0x2718}, // check and ballot marks
}

func narrow(r rune) bool {
	for _, n := range narrowRanges {
		if r < n[0] {
			return false
		}
		if r <= n[1] {
			return true
		}
	}
	return false
}

// textWidth is the width of plain text, without escape sequences.
func textWidth(s string) int {
	w := 0
	for i := 0; i < len(s); {
		c := s[i]
		if c >= 0x20 && c < 0x7f {
			w++
			i++
			continue
		}
		r, n := utf8.DecodeRuneInString(s[i:])
		if !narrow(r) {
			return ansi.StringWidth(s)
		}
		w++
		i += n
	}
	return w
}

// truncateText cuts plain text to at most width columns.
func truncateText(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := 0
	for i := 0; i < len(s); {
		n := 1
		if c := s[i]; c < 0x20 || c >= 0x7f {
			var r rune
			r, n = utf8.DecodeRuneInString(s[i:])
			if !narrow(r) {
				// It may also be a zero width rune that belongs with the one before the cut.
				return ansi.Truncate(s, width, "")
			}
		}
		if w == width {
			return s[:i]
		}
		w++
		i += n
	}
	return s
}

// styledWidth is the width of text with SGR sequences in it, as the views build it.
func styledWidth(s string) int {
	w := 0
	for i := 0; i < len(s); {
		c := s[i]
		switch {
		case c >= 0x20 && c < 0x7f:
			w++
			i++
		case c == 0x1b && i+1 < len(s) && s[i+1] == '[':
			// Parameters and intermediates, up to the final byte.
			j := i + 2
			for j < len(s) && s[j] >= 0x20 && s[j] < 0x40 {
				j++
			}
			if j >= len(s) || s[j] < 0x40 || s[j] >= 0x7f {
				return ansi.StringWidth(s)
			}
			i = j + 1
		default:
			r, n := utf8.DecodeRuneInString(s[i:])
			if !narrow(r) {
				return ansi.StringWidth(s)
			}
			w++
			i += n
		}
	}
	return w
}

// place centres s in an area of width and height, every line on its own as lipgloss.Place does,
// and cuts what does not fit.
func place(width, height int, s string) string {
	lines := strings.Split(s, "\n")
	lines = lines[:min(len(lines), height)]
	blank := strings.Repeat(" ", width)
	top := (height - len(lines)) / 2

	var b strings.Builder
	b.Grow(len(s) + height*width)
	for i := range height {
		if i > 0 {
			b.WriteByte('\n')
		}
		j := i - top
		if j < 0 || j >= len(lines) {
			b.WriteString(blank)
			continue
		}
		line := lines[j]
		w := styledWidth(line)
		if w > width {
			line = ansi.Truncate(line, width, "…")
			w = styledWidth(line)
		}
		gap := width - w
		b.WriteString(blank[:gap/2])
		b.WriteString(line)
		b.WriteString(blank[:gap-gap/2])
	}
	return b.String()
}

// memo keeps renders that only change with their key, like the footer hints and the header, which
// stay the same from one frame to the next while the tables under them change. The palette is part
// of every key, so a cached render never shows the colours of another theme.
type memo struct {
	mu sync.Mutex
	m  map[string]string
}

func (c *memo) get(key string, render func() string) string {
	key = fmt.Sprint(styles.P.Dark) + "\x00" + key
	c.mu.Lock()
	defer c.mu.Unlock()
	if s, ok := c.m[key]; ok {
		return s
	}
	// The keys hold the width and what is on screen, the few that recur stay.
	if c.m == nil || len(c.m) >= 64 {
		c.m = map[string]string{}
	}
	s := render()
	c.m[key] = s
	return s
}
