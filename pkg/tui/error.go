package tui

var (
	errNoPageSelected = "No page selected"
	errNoQueuesFound  = "No queues found"
)

func (m model) ErrorView() string {
	return m.error
}
