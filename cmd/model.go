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
	Cursor int

	Width  int
	Height int

	Keys keyMap
	Help help.Model

	Table table.Model

	State view

	Selected map[int]struct{}
}
