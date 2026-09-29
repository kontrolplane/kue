// Package styles holds the kontrolplane palette as it applies to the terminal.
//
// The values come from the brand tokens (github.com/kontrolplane/assets, tokens/color.json and
// tokens/semantic.json). Terminals cannot blend alpha, so every rule and tint is the solid mix the
// token describes. A detected theme keeps the terminal's own background; both themes paint cornsilk
// and black forest text with olive accents on top of it. A theme chosen with --theme paints its own
// background as well (see Paint).
package styles

import (
	"image/color"
	"strings"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// Palette is one theme of the brand, resolved to solid colours.
type Palette struct {
	Dark bool

	Base      color.Color // the plate or the stock
	Surface   color.Color // raised surfaces, unfocused buttons
	Selection color.Color // the row under the cursor
	Rule      color.Color // hairlines and borders
	RuleBold  color.Color // a rule that has to read as a division

	Text  color.Color // headings and anything that carries the page
	Body  color.Color // values and body copy
	Muted color.Color // labels and meta
	Faint color.Color // hints, punctuation, empty values

	Accent color.Color // olive: the brand's highlight
	Warm   color.Color // copper: numbers in payloads

	Success color.Color
	Warning color.Color
	Info    color.Color
	Danger  color.Color
}

var (
	blackForest = lipgloss.Color("#283618")
	cornsilk    = lipgloss.Color("#FEFAE0")
)

// Dark sits on the terminal's own dark ground. Tints mix cornsilk into a neutral dark so only the
// accent reads as olive.
var Dark = Palette{
	Dark:      true,
	Base:      lipgloss.Color("#1A1A1A"),
	Surface:   lipgloss.Color("#353532"), // 12% cornsilk
	Selection: lipgloss.Color("#3A3F28"), // 45% olive
	Rule:      lipgloss.Color("#3E3E3A"), // 16% cornsilk
	RuleBold:  lipgloss.Color("#716F65"), // 38% cornsilk
	Text:      cornsilk,
	Body:      lipgloss.Color("#ECE8D0"), // 92% cornsilk
	Muted:     lipgloss.Color("#C3C0AD"), // 74% cornsilk
	Faint:     lipgloss.Color("#979587"), // 55% cornsilk
	Accent:    lipgloss.Color("#9FA57B"), // olive lifted for dark grounds
	Warm:      lipgloss.Color("#D0975D"), // copper lifted for the plate
	Success:   lipgloss.Color("#00C389"),
	Warning:   lipgloss.Color("#F7AB1B"),
	Info:      lipgloss.Color("#53A2FF"),
	Danger:    lipgloss.Color("#FF6467"),
}

// Light is the cornsilk stock. Hierarchy is a tint of black forest, never a grey.
var Light = Palette{
	Dark:      false,
	Base:      cornsilk,
	Surface:   lipgloss.Color("#F3F0D6"), // 5% black forest
	Selection: lipgloss.Color("#D7D7BC"), // 18% black forest
	Rule:      lipgloss.Color("#D3D3B8"), // 20% black forest
	RuleBold:  lipgloss.Color("#BEBFA4"), // 30% black forest
	Text:      blackForest,
	Body:      lipgloss.Color("#3D4A2C"), // ink-2
	Muted:     lipgloss.Color("#60694C"), // muted-foreground
	Faint:     lipgloss.Color("#6C7558"), // 68% black forest, 4.5:1 on the stock
	Accent:    lipgloss.Color("#606C38"), // olive
	Warm:      lipgloss.Color("#A16223"), // alert
	Success:   lipgloss.Color("#007E4E"),
	Warning:   lipgloss.Color("#9C6000"),
	Info:      lipgloss.Color("#0067C7"),
	Danger:    lipgloss.Color("#E7000B"),
}

// P is the active palette.
var P = Dark

// Paint is set when the theme was chosen rather than detected. The palette then paints its own
// ground, since the terminal's may be the opposite of what the theme was made for.
var Paint bool

// Use selects the dark or the light theme.
func Use(dark bool) {
	if dark {
		P = Dark
	} else {
		P = Light
	}
}

// Tone names a foreground role, so text can be described once and painted on any ground.
type Tone int

const (
	ToneText Tone = iota
	ToneBody
	ToneMuted
	ToneFaint
	ToneAccent
	ToneWarm
	ToneSuccess
	ToneWarning
	ToneInfo
	ToneDanger
	ToneRule
)

func (p Palette) Color(t Tone) color.Color {
	switch t {
	case ToneBody:
		return p.Body
	case ToneMuted:
		return p.Muted
	case ToneFaint:
		return p.Faint
	case ToneAccent:
		return p.Accent
	case ToneWarm:
		return p.Warm
	case ToneSuccess:
		return p.Success
	case ToneWarning:
		return p.Warning
	case ToneInfo:
		return p.Info
	case ToneDanger:
		return p.Danger
	case ToneRule:
		return p.RuleBold
	default:
		return p.Text
	}
}

// Fg returns a style painting the given tone.
func Fg(t Tone) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(P.Color(t))
}

func Body(s string) string   { return Fg(ToneBody).Render(s) }
func Muted(s string) string  { return Fg(ToneMuted).Render(s) }
func Faint(s string) string  { return Fg(ToneFaint).Render(s) }
func Accent(s string) string { return Fg(ToneAccent).Render(s) }
func Bold(s string) string   { return Fg(ToneText).Bold(true).Render(s) }

// Span is a run of text with a single tone. Cells and inline values are built from spans so they
// can be repainted onto a selected row without the inner resets cutting through its background.
type Span struct {
	Text string
	Tone Tone
	Bold bool
}

func S(text string, tone Tone) Span     { return Span{Text: Clean(text), Tone: tone} }
func B(text string, tone Tone) Span     { return Span{Text: Clean(text), Tone: tone, Bold: true} }
func (s Span) Style() lipgloss.Style    { return Fg(s.Tone).Bold(s.Bold) }
func (s Span) On(bg color.Color) string { return s.Style().Background(bg).Render(s.Text) }

// Render joins spans on the terminal's ground.
func Render(spans ...Span) string {
	var b strings.Builder
	for _, s := range spans {
		b.WriteString(s.Style().Render(s.Text))
	}
	return b.String()
}

// Button renders a dialog button. A focused button takes the primary treatment, or the given
// tone when the action it confirms is destructive.
func Button(label string, focused bool, tone Tone) string {
	style := lipgloss.NewStyle().Padding(0, 3)
	if !focused {
		return style.Foreground(P.Muted).Background(P.Surface).Render(label)
	}
	bg := P.Text
	if tone != ToneText {
		bg = P.Color(tone)
	}
	return style.Foreground(P.Base).Background(bg).Bold(true).Render(label)
}

// Key renders a key binding hint.
func Key(k, desc string) string {
	return Fg(ToneText).Bold(true).Render(k) + " " + Muted(desc)
}

// SectionHeader renders a lowercase label followed by a hairline, filling width.
func SectionHeader(title string, width int, focused bool) string {
	label := Fg(ToneText).Bold(true)
	if focused {
		label = Fg(ToneAccent).Bold(true)
	}
	return SectionHeaderWith(label.Render(title), "", width)
}

// SectionHeaderWith renders a section header with meta text right aligned in the rule.
func SectionHeaderWith(title, meta string, width int) string {
	fill := width - lipgloss.Width(title) - 1
	if meta != "" {
		fill -= lipgloss.Width(meta) + 1
	}
	if fill < 1 {
		fill = 1
	}
	line := title + " " + lipgloss.NewStyle().Foreground(P.Rule).Render(strings.Repeat("─", fill))
	if meta != "" {
		line += " " + meta
	}
	return line
}

// Bar renders a horizontal meter of the given width, filled to ratio.
func Bar(ratio float64, width int, tone Tone) []Span {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	filled := int(ratio*float64(width) + 0.5)
	if ratio > 0 && filled == 0 {
		filled = 1
	}
	return []Span{
		S(strings.Repeat("━", filled), tone),
		S(strings.Repeat("─", width-filled), ToneRule),
	}
}

// TextInput returns the styles for single line inputs.
func TextInput() textinput.Styles {
	s := textinput.DefaultStyles(P.Dark)
	s.Focused.Text = Fg(ToneText)
	s.Focused.Placeholder = Fg(ToneFaint)
	s.Focused.Prompt = Fg(ToneAccent)
	s.Blurred.Text = Fg(ToneBody)
	s.Blurred.Placeholder = Fg(ToneFaint)
	s.Blurred.Prompt = Fg(ToneFaint)
	s.Cursor.Color = P.Accent
	return s
}

// TextArea returns the styles for multi line inputs.
func TextArea() textarea.Styles {
	s := textarea.DefaultStyles(P.Dark)
	for _, state := range []*textarea.StyleState{&s.Focused, &s.Blurred} {
		state.Base = lipgloss.NewStyle()
		state.Text = Fg(ToneBody)
		state.Placeholder = Fg(ToneFaint)
		state.Prompt = lipgloss.NewStyle().Foreground(P.Rule)
		state.LineNumber = Fg(ToneFaint)
		state.CursorLineNumber = Fg(ToneAccent)
		state.CursorLine = lipgloss.NewStyle()
		state.EndOfBuffer = lipgloss.NewStyle().Foreground(P.Base)
	}
	s.Focused.Text = Fg(ToneText)
	s.Focused.Prompt = Fg(ToneAccent)
	s.Focused.CursorLine = lipgloss.NewStyle().Background(P.Surface)
	s.Cursor.Color = P.Accent
	return s
}

// FormTheme returns the huh theme for the queue creation form.
func FormTheme() huh.Theme {
	return huh.ThemeFunc(func(bool) *huh.Styles {
		t := huh.ThemeBase(P.Dark)

		t.Focused.Base = t.Focused.Base.BorderForeground(P.Accent)
		t.Focused.Title = Fg(ToneText).Bold(true)
		t.Focused.Description = Fg(ToneMuted)
		t.Focused.ErrorIndicator = Fg(ToneDanger)
		t.Focused.ErrorMessage = Fg(ToneDanger)
		t.Focused.SelectSelector = Fg(ToneAccent).SetString("› ")
		t.Focused.NextIndicator = Fg(ToneAccent)
		t.Focused.PrevIndicator = Fg(ToneAccent)
		t.Focused.Option = Fg(ToneBody)
		t.Focused.MultiSelectSelector = Fg(ToneAccent).SetString("› ")
		t.Focused.SelectedOption = Fg(ToneText)
		t.Focused.SelectedPrefix = Fg(ToneAccent).SetString("■ ")
		t.Focused.UnselectedOption = Fg(ToneMuted)
		t.Focused.UnselectedPrefix = Fg(ToneFaint).SetString("□ ")
		t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(P.Base).Background(P.Text).Bold(true)
		t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(P.Muted).Background(P.Surface)
		t.Focused.TextInput.Cursor = Fg(ToneAccent)
		t.Focused.TextInput.Placeholder = Fg(ToneFaint)
		t.Focused.TextInput.Prompt = Fg(ToneAccent)
		t.Focused.TextInput.Text = Fg(ToneText)

		t.Blurred = t.Focused
		t.Blurred.Base = t.Blurred.Base.BorderForeground(P.Rule)
		t.Blurred.Title = Fg(ToneMuted)
		t.Blurred.Description = Fg(ToneFaint)
		t.Blurred.TextInput.Prompt = Fg(ToneFaint)
		t.Blurred.TextInput.Text = Fg(ToneBody)
		t.Blurred.NextIndicator = lipgloss.NewStyle()
		t.Blurred.PrevIndicator = lipgloss.NewStyle()

		t.Group.Title = Fg(ToneAccent).Bold(true)
		t.Group.Description = Fg(ToneMuted)
		t.Help.ShortKey = Fg(ToneText).Bold(true)
		t.Help.ShortDesc = Fg(ToneMuted)
		t.Help.ShortSeparator = Fg(ToneFaint)
		t.Help.FullKey = Fg(ToneText).Bold(true)
		t.Help.FullDesc = Fg(ToneMuted)
		t.Help.FullSeparator = Fg(ToneFaint)
		return t
	})
}
