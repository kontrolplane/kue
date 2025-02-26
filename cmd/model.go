package cmd

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/table"
)

type view uint

const (
	Overview view = iota
	Message
)

type model struct {
	cursor   int
	width    int
	height   int
	keys     keyMap
	help     help.Model
	table    table.Model
	state    view
	selected map[int]struct{}
}
