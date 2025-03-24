package tui

type page uint

const (
	queueOverview page = iota
	queueDetails
	queueCreate
	queueDelete
	queueMessageDetails
	queueMessageCreate
	queueMessageDelete
)

func (m model) SwitchPage(page page) model {
	m.page = page
	return m
}
