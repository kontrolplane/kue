package tui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/spinner"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"github.com/kontrolplane/kue/pkg/client"
	keys "github.com/kontrolplane/kue/pkg/keys"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// Layout constants for consistent sizing across all views
const (
	attributesHeight = 7   // Attributes plus spacing in the details view
	tableHeaderRows  = 2   // Column titles and the rule under them
	minTableHeight   = 5   // Minimum body rows for any table
	minContentWidth  = 100 // Below this, or minContentHeight, a notice asks for a larger terminal
	minContentHeight = 16
	chromeWidth      = 2 // Frame borders
	chromeHeight     = 9 // Header, spacing, frame edges and padding, footer
)

// The content area fills the terminal. setLayout resizes it and everything derived from it.
var (
	contentWidth  = 140
	contentHeight = 25
	frameWidth    = contentWidth + chromeWidth
)

func setLayout(width, height int) {
	contentWidth = max(minContentWidth, width-chromeWidth)
	contentHeight = max(minContentHeight, height-chromeHeight)
	frameWidth = contentWidth + chromeWidth
	setPanelLayout()
}

// resize fits the tables, viewports and inputs of every page to the current layout.
func (m model) resize() model {
	m.state.queueOverview.table.setSize(contentWidth-4, contentHeight-tableHeaderRows)

	d := &m.state.queueDetails
	d.messagesTable.setSize(contentWidth-4, m.getMessageTableHeight())

	// The viewport and inputs are built when their page opens, so only the open one needs fitting.
	switch m.page {
	case queueMessageDetails:
		m.state.queueMessageDetails.resizeBody()
	case queueMessageCreate:
		c := &m.state.queueMessageCreate
		c.textarea.SetWidth(rightContentWidth)
		c.textarea.SetHeight(messageBodyHeight())
	case queueCreate:
		if f := m.state.queueCreate.form; f != nil {
			f.WithWidth(formWidth()).WithHeight(formHeight())
		}
	case queueDelete:
		c := &m.state.queueDelete.confirm
		c.input.SetWidth(typedConfirmWidth(c.want))
	}
	return m
}

type model struct {
	projectName string
	programName string
	page        page
	state       state
	client      *sqs.Client
	awsInfo     client.AWSInfo
	context     context.Context
	width       int
	height      int
	keys        keys.KeyMap
	showHelp    bool
	error       string
	loading     bool
	loadingMsg  string
	spinner     spinner.Model
	spinning    bool
	statusMsg   string
	statusTone  styles.Tone
	statusGen   int
	refreshGen  int
	busy        bool // a create, delete, purge, redrive or send is in flight
	armed       bool // esc was pressed once on a form with input, the next esc discards it
	paused      bool // the refresh ticks are stopped
	refreshedAt time.Time
	clocking    bool // a tick for the time since the last refresh is pending
}

// getMessageTableHeight returns the body rows of the messages table in the details view.
func (m model) getMessageTableHeight() int {
	return max(minTableHeight, contentHeight-attributesHeight-tableHeaderRows)
}

type state struct {
	queueOverview       queueOverviewState
	queueDetails        queueDetailsState
	queueDelete         queueDeleteState
	queueCreate         queueCreateState
	queuePurge          queuePurgeState
	queueRedrive        queueRedriveState
	queueMessageDetails queueMessageDetailsState
	queueMessageCreate  queueMessageCreateState
	queueMessageDelete  queueMessageDeleteState
}
