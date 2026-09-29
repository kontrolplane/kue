package keys

import "charm.land/bubbles/v2/key"

// KeyMap holds the keybindings shared by the views. Moving through a table (up/down, j/k,
// pgup/pgdown) is handled by the table itself.
type KeyMap struct {
	Left            key.Binding
	Right           key.Binding
	Top             key.Binding
	Bottom          key.Binding
	Help            key.Binding
	View            key.Binding
	Select          key.Binding
	Filter          key.Binding
	Create          key.Binding
	Delete          key.Binding
	CopyToClipboard key.Binding
	Purge           key.Binding
	Redrive         key.Binding
	Refresh         key.Binding
	Pause           key.Binding
	NextField       key.Binding
	PrevField       key.Binding
	Submit          key.Binding
	Yes             key.Binding
	No              key.Binding
	Back            key.Binding
	Quit            key.Binding
	ForceQuit       key.Binding
}

var Keys = KeyMap{
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("← | h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→ | l", "right"),
	),
	Top: key.NewBinding(
		key.WithKeys("g", "home"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G", "end"),
		key.WithHelp("G", "bottom"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	View: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "view"),
	),
	Select: key.NewBinding(
		key.WithKeys("space"),
		key.WithHelp("space", "select"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Create: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "create"),
	),
	Delete: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "delete"),
	),
	CopyToClipboard: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "copy"),
	),
	Purge: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "purge"),
	),
	Redrive: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "redrive"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Pause: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "pause refresh"),
	),
	NextField: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next field"),
	),
	PrevField: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "previous field"),
	),
	Submit: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "submit"),
	),
	Yes: key.NewBinding(
		key.WithKeys("y"),
		key.WithHelp("y", "yes"),
	),
	No: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "no"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	ForceQuit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}
