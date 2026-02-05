// Package styles provides centralized styling for the TUI application.
package styles

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Color palette
var (
	// Primary brand color (sage green)
	AccentColor = lipgloss.Color("#628049")

	// Text colors
	TextWhite = lipgloss.Color("#ffffff")
	TextLight = lipgloss.Color("255")

	// Background/border colors
	BorderColor    = lipgloss.Color("240")
	DarkGray       = lipgloss.Color("240")
	MediumGray     = lipgloss.Color("243")
	LightGray      = lipgloss.Color("250")
	NearWhite      = lipgloss.Color("255")
)

// MainBorder is the standard border style for main content areas.
var MainBorder = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(BorderColor).
	Padding(1, 0)

// ButtonPrimary is the style for primary (unfocused/default) buttons.
var ButtonPrimary = lipgloss.NewStyle().
	Foreground(NearWhite).
	Background(DarkGray).
	Padding(0, 3)

// ButtonSecondary is the style for secondary (focused/active) buttons.
var ButtonSecondary = lipgloss.NewStyle().
	Foreground(TextWhite).
	Background(AccentColor).
	Padding(0, 3)

// DialogContainer is the style for dialog boxes.
var DialogContainer = lipgloss.NewStyle().
	Padding(1, 3)

// Bold returns a bold text style.
var Bold = lipgloss.NewStyle().Bold(true)

// ContentWrapper creates a centered content wrapper with the given dimensions.
func ContentWrapper(width, height int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)
}

// CenteredForm creates a centered form wrapper with the given width.
func CenteredForm(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width - 4).
		Align(lipgloss.Center).
		PaddingTop(2)
}

// TableStyles returns the standard styles used for tables in the TUI.
func TableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(BorderColor).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(TextWhite).
		Background(AccentColor).
		Bold(false)
	return s
}

// AttributesTableStyles returns styles for the attributes display table (no selection highlight).
func AttributesTableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.HiddenBorder())
	s.Selected = s.Selected.
		Foreground(lipgloss.NoColor{}).
		Bold(false)
	return s
}

// FormTheme returns the huh form theme with brand colors.
func FormTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Title = t.Focused.Title.Foreground(AccentColor).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(MediumGray)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(AccentColor)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Background(AccentColor).Foreground(TextWhite)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Background(DarkGray).Foreground(NearWhite)

	t.Blurred.Title = t.Blurred.Title.Foreground(LightGray)
	t.Blurred.Description = t.Blurred.Description.Foreground(DarkGray)

	return t
}
