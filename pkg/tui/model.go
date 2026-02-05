package tui

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/charmbracelet/bubbles/help"

	"github.com/kontrolplane/kue/pkg/client"
	keys "github.com/kontrolplane/kue/pkg/keys"
)

// Layout constants for consistent sizing across all views
const (
	headerHeight          = 3   // Height of the header section
	footerPadding         = 4   // Padding for footer/help area
	borderPadding         = 2   // Border adds 2 chars (left + right or top + bottom)
	attributesTableHeight = 8   // Reserved height for attributes table in details view
	minTableHeight        = 5   // Minimum height for any table
	defaultTableHeight    = 10  // Default table height when window size unknown
	contentWidth          = 140 // Fixed width for content area
	contentHeight         = 25  // Fixed height for content area
)

type model struct {
	projectName string
	programName string
	page        page
	previous    page
	state       state
	client      *sqs.Client
	awsInfo     client.AWSInfo
	context     context.Context
	width       int
	height      int
	keys        keys.KeyMap
	help        help.Model
	error       string
	loading     bool
	loadingMsg  string
}

// getTableHeight returns the height available for tables.
func (m model) getTableHeight() int {
	// Account for table header (2 lines) and some padding
	return contentHeight - 3
}

// getMessageTableHeight returns the height for the messages table in details view.
func (m model) getMessageTableHeight() int {
	// Content height minus attributes table area
	available := contentHeight - attributesTableHeight - 3
	if available < minTableHeight {
		return minTableHeight
	}
	return available
}

type state struct {
	queueOverview         queueOverviewState
	queueDetails          queueDetailsState
	queueDelete           queueDeleteState
	queueCreate           queueCreateState
	queueMessageDetails   queueMessageDetailsState
	queueMessageCreate    queueMessageCreateState
	queueMessageDelete    queueMessageDeleteState
}
