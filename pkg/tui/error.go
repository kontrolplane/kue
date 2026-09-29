package tui

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

var errNoPageSelected = "no page selected"

func (m model) ErrorView() string {
	text := styles.CleanBlock(m.error)
	width := min(80, dialogTextWidth(), lipgloss.Width(text))
	lines := strings.Split(styles.Fg(styles.ToneBody).Width(width).Render(text), "\n")
	// Room for the heading, the hint below and the card's edges.
	if room := max(1, contentHeight-6); len(lines) > room {
		lines = append(lines[:room-1], styles.Faint("…"))
	}
	return dialog("error", styles.ToneDanger, strings.Join(lines, "\n"), "", styles.Faint("press any key to continue"))
}
