package tui

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/charmbracelet/bubbles/help"

	tea "github.com/charmbracelet/bubbletea"
	keys "github.com/kontrolplane/kue/pkg/keys"
)

type model struct {
	projectName string
	programName string
	page        page
	previous    page
	state       state
	client      *sqs.Client
	context     context.Context
	width       int
	height      int
	keys        keys.KeyMap
	help        help.Model
	table       tea.Model
	error       string
	loading     bool
	loadingMsg  string
}

type state struct {
	queueOverview      queueOverviewState
	queueDetails       queueDetailsState
	queueDelete        queueDeleteState
	queueCreate        queueCreateState
	queueMessageCreate queueMessageCreateState
}
