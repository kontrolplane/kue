package tui

import "time"

type page uint

const (
	queueOverview page = iota
	queueDetails
	queueCreate
	queueDelete
	queueRedrive
	queuePurge
	queueMessageDetails
	queueMessageCreate
	queueMessageDelete
)

// discardPrompt asks to confirm leaving a form with input. It belongs to the form, and goes with it.
const discardPrompt = "press esc again to discard"

// SwitchPage opens page. A load still running for the page left behind no longer holds the
// spinner, a page that loads its own data starts it again.
func (m model) SwitchPage(page page) model {
	if page != m.page {
		m.refreshedAt = time.Time{}
		m.armed = false
		if m.statusMsg == discardPrompt {
			m.statusMsg = ""
		}
	}
	m.page = page
	if !m.busy {
		m.loading = false
	}
	return m
}
