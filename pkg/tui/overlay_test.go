package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestOverlayKeepsThePageAround(t *testing.T) {
	t.Cleanup(func() { setLayout(142, 34) })
	setLayout(minContentWidth+chromeWidth, minContentHeight+chromeHeight)
	page := make([]string, contentHeight)
	for i := range page {
		page[i] = strings.Repeat(string(rune('a'+i)), contentWidth)
	}
	got := strings.Split(ansi.Strip(overlay(strings.Join(page, "\n"), "XX")), "\n")
	mid := (contentHeight - 1) / 2
	for i, line := range got {
		want := page[i]
		if i == mid {
			// The card, with a blank cell either side of it.
			at := (contentWidth - 4) / 2
			want = want[:at] + " XX " + want[at+4:]
		}
		if line != want {
			t.Errorf("line %d = %q, want %q", i, line, want)
		}
	}
}

func TestDialogsShowTheirPage(t *testing.T) {
	tests := []struct {
		name string
		keys []string
		page page
		want []string // on the page behind the dialog
	}{
		{"purge from the queue", []string{"ctrl+p"}, queuePurge, []string{"11111111-aaaa", "purge queue"}},
		{"delete a message", []string{"ctrl+d"}, queueMessageDelete, []string{"22222222-bbbb", "delete message"}},
		{"delete from its details", []string{"enter", "ctrl+d"}, queueMessageDelete, []string{"source", "delete message"}},
		{"delete a queue", []string{"q", "ctrl+d"}, queueDelete, []string{"payments.fifo", "delete queue"}},
	}
	for _, tt := range tests {
		m := detailsModel(t)
		for _, k := range tt.keys {
			m, _ = update(t, m, press(k))
		}
		if m.page != tt.page {
			t.Fatalf("%s: page %v, want %v", tt.name, m.page, tt.page)
		}
		view := ansi.Strip(m.render())
		for _, want := range tt.want {
			if !strings.Contains(view, want) {
				t.Errorf("%s: expected %q on screen", tt.name, want)
			}
		}
	}

	m := detailsModel(t)
	m.error = "boom"
	if view := ansi.Strip(m.render()); !strings.Contains(view, "boom") || !strings.Contains(view, "11111111-aaaa") {
		t.Error("expected the error over the queue it happened on")
	}
}

func TestHelpStartsWithThePage(t *testing.T) {
	for _, tt := range []struct {
		keys  []string
		first string
	}{
		{[]string{"q"}, "queues"},
		{nil, "queue "},
		{[]string{"enter"}, "message"},
		{[]string{"ctrl+p"}, "forms and dialogs"},
	} {
		m := detailsModel(t)
		for _, k := range append(tt.keys, "?") {
			m, _ = update(t, m, press(k))
		}
		help := ansi.Strip(m.renderHelp())
		var titles []string
		for _, s := range helpSections {
			titles = append(titles, s.title)
		}
		first := len(help)
		for _, title := range titles {
			if i := strings.Index(help, title+" "); i >= 0 && i < first {
				first = i
			}
		}
		if got := strings.Index(help, tt.first); got != first {
			t.Errorf("keys %v: expected the help to start with %q:\n%s", tt.keys, tt.first, help)
		}
	}
}

// Over a page with blank lines, as an empty table has, the card still sits in the middle.
func TestOverlayCentresOverBlankLines(t *testing.T) {
	t.Cleanup(func() { setLayout(142, 34) })
	setLayout(minContentWidth+chromeWidth, minContentHeight+chromeHeight)
	got := strings.Split(ansi.Strip(overlay("short", "XX")), "\n")
	mid := (contentHeight - 1) / 2
	if at := strings.Index(got[mid], "XX"); at != (contentWidth-4)/2+1 {
		t.Errorf("card at column %d, want %d: %q", at, (contentWidth-4)/2+1, got[mid])
	}
	for i, line := range got {
		if w := ansi.StringWidth(line); w != contentWidth {
			t.Errorf("line %d is %d wide, want the content width %d", i, w, contentWidth)
		}
	}
}
