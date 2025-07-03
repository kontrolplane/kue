package keys

import "github.com/charmbracelet/bubbles/key"

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map. It could also very easily be a map[string]key.Binding.
type KeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Left   key.Binding
	Right  key.Binding
	Help   key.Binding
	View   key.Binding
	Select key.Binding
	Filter key.Binding
	Create key.Binding
	Delete key.Binding
	Quit   key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.View, k.Help, k.Filter, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{ // first column
			k.Up,
			k.Down,
			k.Left,
			k.Right,
		},
		{ // second column
			k.View,
			k.Select,
			k.Help,
			k.Filter,
			k.Create,
			k.Delete,
			k.Quit,
		},
	}
}

var Keys = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑ | k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓ | j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("← | h", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→ | l", "right"),
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
		key.WithKeys(" "),
		key.WithHelp("space", "select"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
	Create: key.NewBinding(
		key.WithKeys("c", "ctrl+n"),
		key.WithHelp("c", "create"),
	),
	Delete: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "delete"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
