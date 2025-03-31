package tui

type page uint

const (
	queueOverview page = iota
	queueDetails
	queueCreate
	queueDelete
	queueRedrive
	queueMessageDetails
	queueMessageCreate
	queueMessageDelete
)

var views = map[page]string{
	queueOverview:       "queue overview",
	queueDetails:        "queue details",
	queueCreate:         "queue create",
	queueDelete:         "queue delete",
	queueRedrive:        "queue redrive",
	queueMessageDetails: "queue message details",
	queueMessageCreate:  "queue message create",
	queueMessageDelete:  "queue message delete",
}

func (m model) SwitchPage(page page) model {
	m.page = page
	return m
}
