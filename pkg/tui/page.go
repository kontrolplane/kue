package tui

func (m model) SwitchPage(page page) model {
	m.page = page
	return m
}