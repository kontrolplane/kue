package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

type confirmResult int

const (
	confirmPending confirmResult = iota
	confirmYes
	confirmNo
)

// confirmKey handles the no/yes choice of a dialog, toggling selected in place. Once answered,
// selected is back on "no" for the next time the dialog opens.
func (m model) confirmKey(msg tea.Msg, selected *int) confirmResult {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return confirmPending
	}
	result := confirmPending
	switch {
	case key.Matches(keyMsg, m.keys.Left, m.keys.Right, m.keys.NextField, m.keys.PrevField):
		*selected = 1 - *selected
	case key.Matches(keyMsg, m.keys.Yes):
		result = confirmYes
	case key.Matches(keyMsg, m.keys.No, m.keys.Quit, m.keys.Back):
		result = confirmNo
	case key.Matches(keyMsg, m.keys.View):
		result = confirmNo
		if *selected == 1 {
			result = confirmYes
		}
	}
	if result != confirmPending {
		*selected = 0
	}
	return result
}

// confirmButtons renders a no/yes button pair. selected 0 highlights "no", 1 highlights "yes"
// and any other value highlights neither.
func confirmButtons(selected int, tone styles.Tone) string {
	return lipgloss.JoinHorizontal(lipgloss.Center,
		styles.Button("no", selected == 0, styles.ToneText),
		"    ",
		styles.Button("yes", selected == 1, tone),
	)
}

// dialogPadX is the room a dialog card leaves between its edge and its text.
const dialogPadX = 4

// dialogTextWidth is the widest a line of dialog text gets before it wraps.
func dialogTextWidth() int {
	return min(contentWidth-2-2*dialogPadX, 100)
}

// dialogName bounds a name set into a dialog sentence, so the sentence stays readable.
func dialogName(name string) string {
	return truncate(name, 48)
}

// dialog renders a card with a toned edge, used for confirmations and errors, to be set over a
// page with overlay. Lines wider than the card wraps, and the card gives up its inner margin when
// it would not fit the height.
func dialog(title string, tone styles.Tone, body ...string) string {
	heading := styles.Render(styles.B("▲ ", tone), styles.B(title, tone))
	width := dialogTextWidth()
	lines := []string{heading, ""}
	for _, block := range body {
		for _, line := range strings.Split(block, "\n") {
			if lipgloss.Width(line) > width {
				line = lipgloss.Wrap(line, width, "")
			}
			lines = append(lines, line)
		}
	}
	content := lipgloss.JoinVertical(lipgloss.Center, lines...)
	padY := 1
	if lipgloss.Height(content)+2+2*padY > contentHeight {
		padY = 0
	}
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.P.Color(tone)).
		Padding(padY, dialogPadX).
		Render(content)
	return card
}

// overlay sets card in the middle of the content area, over the page, which recedes behind it, so
// a dialog keeps what it acts on in sight.
func overlay(page, card string) string {
	lines := strings.Split(ansi.Strip(page), "\n")
	dim := sgr(styles.P.Rule, false, nil)
	for i, l := range lines {
		lines[i] = dim + l + ansi.ResetStyle
	}
	// A ring of blank cells keeps the page's text off the card's edge.
	card = lipgloss.NewStyle().Padding(0, 1).Render(card)
	w, h := lipgloss.Width(card), lipgloss.Height(card)
	c := lipgloss.NewCanvas(contentWidth, contentHeight)
	c.Compose(lipgloss.NewCompositor(
		lipgloss.NewLayer(strings.Join(lines, "\n")),
		lipgloss.NewLayer(card).X(max(0, (contentWidth-w)/2)).Y(max(0, (contentHeight-h)/2)).Z(1),
	))
	// The canvas leaves out the blank cells at the end of a line, and the frame centres a line
	// shorter than the content area, which would move the card off the middle.
	out := strings.Split(c.Render(), "\n")
	for i, l := range out {
		out[i] = l + strings.Repeat(" ", max(0, contentWidth-styledWidth(l)))
	}
	return strings.Join(out, "\n")
}

// confirmDialog asks a destructive yes/no question: the prompt, what it affects, and the buttons.
func confirmDialog(title string, prompt []styles.Span, body []string, selected int) string {
	lines := append([]string{styles.Render(prompt...)}, body...)
	return dialog(title, styles.ToneDanger, append(lines, "", confirmButtons(selected, styles.ToneDanger))...)
}

// maxListed bounds the items a dialog names before it counts the rest.
const maxListed = 8

// listed is how many items a dialog names, one per line, before it counts the rest. It leaves
// room for the rest of the dialog in the content height.
func listed() int {
	return max(1, min(maxListed, contentHeight-15))
}

// itemList names the items a dialog acts on, on one line each or separated by commas, with the
// ones past what fits counted.
func itemList(items []string, oneline bool) string {
	limit := maxListed
	if !oneline {
		limit = listed()
	}
	shown := items
	if len(items) > limit {
		shown = items[:limit]
	}
	lines := make([]string, len(shown))
	for i, item := range shown {
		lines[i] = styles.Render(styles.S(truncate(item, dialogTextWidth()), styles.ToneText))
	}
	sep := "\n"
	if oneline {
		sep = styles.Faint(", ")
	}
	list := strings.Join(lines, sep)
	if more := len(items) - len(shown); more > 0 {
		if oneline {
			list += styles.Faint(", ")
		} else {
			list += "\n"
		}
		list += styles.Faint(fmt.Sprintf("… +%d more", more))
	}
	return list
}

// listBody names the items of a dialog when there is more than one, the prompt names a single one.
func listBody(items []string, oneline bool) []string {
	if len(items) < 2 {
		return nil
	}
	return []string{"", itemList(items, oneline)}
}

func deletingMsg(n int, noun string) string {
	if n > 1 {
		return fmt.Sprintf("deleting %d %ss…", n, noun)
	}
	return "deleting " + noun + "…"
}

// typedConfirm guards an action that cannot be undone behind typing a phrase, usually the name of
// what it removes.
type typedConfirm struct {
	want  string
	input textinput.Model
}

func newTypedConfirm(want string) (typedConfirm, tea.Cmd) {
	input := textinput.New()
	input.CharLimit = 255
	input.SetWidth(typedConfirmWidth(want))
	input.Prompt = "› "
	input.SetStyles(styles.TextInput())
	cmd := input.Focus()
	return typedConfirm{want: want, input: input}, cmd
}

// typedConfirmWidth fits the input to the phrase, within the dialog.
func typedConfirmWidth(want string) int {
	return min(max(24, ansi.StringWidth(want)+2), dialogTextWidth()-6)
}

func (c typedConfirm) ok() bool {
	return strings.TrimSpace(c.input.Value()) == c.want
}

func (c typedConfirm) update(msg tea.Msg) (typedConfirm, tea.Cmd) {
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	return c, cmd
}

// view asks for the phrase, and marks whether what was typed matches it.
func (c typedConfirm) view() string {
	mark := "  "
	switch {
	case c.ok():
		mark = styles.Render(styles.S("✓ ", styles.ToneSuccess))
	case c.input.Value() != "":
		mark = styles.Render(styles.S("✗ ", styles.ToneDanger))
	}
	ask := styles.Render(styles.S("type ", styles.ToneFaint), styles.B(c.want, styles.ToneText), styles.S(" to confirm", styles.ToneFaint))
	return lipgloss.JoinVertical(lipgloss.Center, lipgloss.Wrap(ask, dialogTextWidth(), ""), "", mark+c.input.View())
}
