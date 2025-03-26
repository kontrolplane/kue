package tui

import (
	"context"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	keys "github.com/kontrolplane/kue/pkg/keys"
)

type model struct {
	projectName string
	programName string
	viewName    string
	page        page
	previous    page
	state       state
	context     context.Context
	width       int
	height      int
	keys        keys.KeyMap
	help        help.Model
	table       tea.Model
	error       string
}

type state struct {
	queueOverview      queueOverviewState
	queueDetails       queueDetailsState
	queueDelete        queueDeleteState
	queueCreate        queueCreateState
	queueMessageCreate queueMessageCreateState
}
