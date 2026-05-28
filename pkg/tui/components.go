package tui

import (
	"charm.land/lipgloss/v2"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// renderConfirmButtons renders a pair of no/yes buttons with the selected one highlighted.
// selected: 0 = no highlighted, 1 = yes highlighted.
func renderConfirmButtons(selected int) string {
	confirm := "yes"
	abort := "no"

	if selected == 0 {
		abort = styles.ButtonSecondary.Render(abort)
		confirm = styles.ButtonPrimary.Render(confirm)
	} else {
		abort = styles.ButtonPrimary.Render(abort)
		confirm = styles.ButtonSecondary.Render(confirm)
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, abort, "    ", confirm)
}

// renderConfirmDialog renders a centered confirmation dialog with a title, prompt lines, and buttons.
func renderConfirmDialog(title string, selected int, promptLines ...string) string {
	lines := []string{title, ""}
	lines = append(lines, promptLines...)
	lines = append(lines, "", renderConfirmButtons(selected))

	dialog := lipgloss.JoinVertical(lipgloss.Center, lines...)
	return lipgloss.Place(contentWidth, contentHeight-2, lipgloss.Center, lipgloss.Center, dialog)
}

// renderVerticalDivider renders a vertical line of the given height using the border color.
func renderVerticalDivider(height int) string {
	var lines string
	for i := 0; i < height; i++ {
		lines += "│"
		if i < height-1 {
			lines += "\n"
		}
	}
	return lipgloss.NewStyle().Foreground(styles.BorderColor).Render(lines)
}
