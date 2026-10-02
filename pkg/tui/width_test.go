package tui

import (
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

func TestWidthsMatchAnsi(t *testing.T) {
	plain := []string{"", "GET /orders 200", "naïve café", "日本語のログ", "emoji 🚀 ok", `\x1b escaped`, "tab\tin",
		"e\u0301 combined", "╭─ box ─╮ … ▌ ● ✓", "flag 🇳🇱", "♥\ufe0f"}
	for _, s := range plain {
		if got, want := textWidth(s), ansi.StringWidth(s); got != want {
			t.Errorf("textWidth(%q) = %d, want %d", s, got, want)
		}
		for w := range ansi.StringWidth(s) + 2 {
			if got, want := truncateText(s, w), ansi.Truncate(s, w, ""); got != want {
				t.Errorf("truncateText(%q, %d) = %q, want %q", s, w, got, want)
			}
		}
	}
	styled := append(plain,
		styles.Render(styles.B("bold", styles.ToneAccent), styles.S(" and plain", styles.ToneFaint)),
		pad(4, styles.P.Selection)+styles.S("日本", styles.ToneText).On(styles.P.Selection),
		"\x1b]8;;https://example.com\x07link\x1b]8;;\x07",
		"\x1b[38;2;1;2;3mtruecolor\x1b[m",
		"\x1b[",
	)
	for _, s := range styled {
		if got, want := styledWidth(s), ansi.StringWidth(s); got != want {
			t.Errorf("styledWidth(%q) = %d, want %d", s, got, want)
		}
	}
}

func TestNarrowRunesAreOneColumn(t *testing.T) {
	for _, n := range narrowRanges {
		for r := n[0]; r <= n[1]; r++ {
			if w := ansi.StringWidth(string(r)); w != 1 {
				t.Errorf("%U %q is %d wide", r, r, w)
			}
		}
	}
}

func TestPlaceMatchesLipgloss(t *testing.T) {
	block := "short\n" + styles.Render(styles.B("a longer line │ with ▌ symbols", styles.ToneAccent)) + "\n\nend"
	for _, size := range [][2]int{{40, 10}, {41, 9}, {33, 4}} {
		got := place(size[0], size[1], block)
		want := lipgloss.Place(size[0], size[1], lipgloss.Center, lipgloss.Center, block)
		if got != want {
			t.Errorf("%v:\n%q\nwant\n%q", size, got, want)
		}
	}
}
