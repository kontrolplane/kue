package tui

import (
	"strings"

	"charm.land/bubbles/v2/textinput"

	tea "charm.land/bubbletea/v2"
)

// filterModel encapsulates text-based filtering state used by list pages.
type filterModel struct {
	active bool
	input  textinput.Model
	text   string
}

// newFilter creates a new filter model with the given placeholder text.
func newFilter(placeholder string) filterModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 50
	ti.SetWidth(30)
	return filterModel{input: ti}
}

// Update handles key events while the filter is active.
// Returns the updated filter and a command. When the filter is not active,
// this is a no-op.
func (f filterModel) Update(msg tea.Msg) (filterModel, tea.Cmd) {
	if !f.active {
		return f, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.Code {
		case tea.KeyEscape:
			f.active = false
			f.input.Blur()
			return f, nil
		case tea.KeyEnter:
			f.active = false
			f.text = f.input.Value()
			f.input.Blur()
			return f, nil
		}
	}

	var cmd tea.Cmd
	f.input, cmd = f.input.Update(msg)
	// Live filtering as user types
	f.text = f.input.Value()
	return f, cmd
}

// Activate enters filter mode and focuses the text input.
func (f filterModel) Activate() (filterModel, tea.Cmd) {
	f.active = true
	f.input.Focus()
	return f, textinput.Blink
}

// Clear removes the current filter text without deactivating.
func (f filterModel) Clear() filterModel {
	f.text = ""
	f.input.SetValue("")
	return f
}

// Reset removes all filter state and deactivates.
func (f filterModel) Reset() filterModel {
	f.active = false
	f.text = ""
	f.input.SetValue("")
	f.input.Blur()
	return f
}

// Matches returns true if s contains the filter text (case-insensitive).
// Returns true when there is no active filter text.
func (f filterModel) Matches(s string) bool {
	if f.text == "" {
		return true
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(f.text))
}

// View returns the rendered filter input.
func (f filterModel) View() string {
	return f.input.View()
}
