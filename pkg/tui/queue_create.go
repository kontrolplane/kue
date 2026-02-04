package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

var queueNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type queueCreateInput struct {
	name                      string
	queueType                 string
	visibilityTimeout         string
	messageRetentionPeriod    string
	deliveryDelay             string
	maximumMessageSize        string
	receiveMessageWaitTime    string
	contentBasedDeduplication bool
	deduplicationScope        string
	fifoThroughputLimit       string
}

type queueCreateState struct {
	input *queueCreateInput
	form  *huh.Form
}

func newQueueCreateForm(input *queueCreateInput) *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Queue Name").
				Description("Alphanumeric characters, hyphens, and underscores only (1-80 chars)").
				Placeholder("queue-name").
				Value(&input.name).
				Validate(func(s string) error {
					s = strings.TrimSpace(s)
					if s == "" {
						return fmt.Errorf("queue name is required")
					}
					if len(s) > 80 {
						return fmt.Errorf("queue name must be 80 characters or less")
					}
					if !queueNameRegex.MatchString(s) {
						return fmt.Errorf("only alphanumeric characters, hyphens, and underscores allowed")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Queue Type").
				Description("Standard queues offer best-effort ordering. FIFO queues guarantee ordering.").
				Options(
					huh.NewOption("Standard", "standard"),
					huh.NewOption("FIFO", "fifo"),
				).
				Value(&input.queueType),
		).Title("Basic Configuration"),

		huh.NewGroup(
			huh.NewInput().
				Title("Visibility Timeout (seconds)").
				Description("Duration a message is hidden after being received (0-43200, default: 30)").
				Placeholder("30").
				Value(&input.visibilityTimeout).
				Validate(validateIntRange(0, 43200)),

			huh.NewInput().
				Title("Message Retention Period (seconds)").
				Description("How long messages are kept (60-1209600, default: 345600 = 4 days)").
				Placeholder("345600").
				Value(&input.messageRetentionPeriod).
				Validate(validateIntRangeOrEmpty(60, 1209600)),

			huh.NewInput().
				Title("Delivery Delay (seconds)").
				Description("Delay before messages become visible (0-900, default: 0)").
				Placeholder("0").
				Value(&input.deliveryDelay).
				Validate(validateIntRange(0, 900)),
		).Title("Message Settings"),

		huh.NewGroup(
			huh.NewInput().
				Title("Maximum Message Size (bytes)").
				Description("Maximum size of a message (1024-262144, default: 262144 = 256KB)").
				Placeholder("262144").
				Value(&input.maximumMessageSize).
				Validate(validateIntRangeOrEmpty(1024, 262144)),

			huh.NewInput().
				Title("Receive Message Wait Time (seconds)").
				Description("Long polling wait time (0-20, default: 0)").
				Placeholder("0").
				Value(&input.receiveMessageWaitTime).
				Validate(validateIntRange(0, 20)),
		).Title("Advanced Settings"),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Content-Based Deduplication").
				Description("Enable automatic deduplication based on message body (FIFO only)").
				Value(&input.contentBasedDeduplication),

			huh.NewSelect[string]().
				Title("Deduplication Scope").
				Description("Scope of message deduplication (FIFO only)").
				Options(
					huh.NewOption("Queue (default)", "queue"),
					huh.NewOption("Message Group", "messageGroup"),
				).
				Value(&input.deduplicationScope),

			huh.NewSelect[string]().
				Title("FIFO Throughput Limit").
				Description("Throughput limit for FIFO queues").
				Options(
					huh.NewOption("Per Queue (default)", "perQueue"),
					huh.NewOption("Per Message Group ID", "perMessageGroupId"),
				).
				Value(&input.fifoThroughputLimit),
		).Title("FIFO Settings (only applies to FIFO queues)"),
	).
		WithTheme(styles.FormTheme()).
		WithShowHelp(false).
		WithWidth(80)

	return form
}

func validateIntRange(min, max int) func(string) error {
	return func(s string) error {
		if s == "" {
			return nil
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a valid number")
		}
		if val < min || val > max {
			return fmt.Errorf("must be between %d and %d", min, max)
		}
		return nil
	}
}

func validateIntRangeOrEmpty(min, max int) func(string) error {
	return func(s string) error {
		if s == "" {
			return nil
		}
		val, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("must be a valid number")
		}
		if val < min || val > max {
			return fmt.Errorf("must be between %d and %d", min, max)
		}
		return nil
	}
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	// Clear any previous error
	m.error = ""

	// Initialize with defaults
	m.state.queueCreate.input = &queueCreateInput{
		queueType:           "standard",
		deduplicationScope:  "queue",
		fifoThroughputLimit: "perQueue",
	}

	m.state.queueCreate.form = newQueueCreateForm(m.state.queueCreate.input)

	return m.SwitchPage(queueCreate), m.state.queueCreate.form.Init()
}

func (m model) QueueCreateView() string {
	if m.state.queueCreate.form == nil {
		return "Loading..."
	}

	formView := m.state.queueCreate.form.View()

	// Center the form both horizontally and vertically
	return lipgloss.Place(contentWidth, contentHeight,
		lipgloss.Center, lipgloss.Center,
		formView)
}

func (m model) QueueCreateUpdate(msg tea.Msg) (model, tea.Cmd) {
	if m.state.queueCreate.form == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Only use esc to cancel on form pages (not 'q' since user needs to type)
		if msg.String() == "esc" {
			return m.QueueOverviewSwitchPage(msg)
		}
	}

	// Update the form
	form, cmd := m.state.queueCreate.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.state.queueCreate.form = f
	}

	// Check if form is completed
	if m.state.queueCreate.form.State == huh.StateCompleted {
		input := m.state.queueCreate.input

		queueName := strings.TrimSpace(input.name)

		// Validate queue name before API call
		if queueName == "" {
			m.error = "Queue name is required"
			return m.QueueOverviewSwitchPage(msg)
		}
		if !queueNameRegex.MatchString(queueName) {
			m.error = "Queue name can only contain alphanumeric characters, hyphens, and underscores"
			return m.QueueOverviewSwitchPage(msg)
		}

		config := kue.QueueConfig{
			Name:   queueName,
			IsFifo: input.queueType == "fifo",
		}

		// Parse optional integer fields
		if input.visibilityTimeout != "" {
			if val, err := strconv.Atoi(input.visibilityTimeout); err == nil {
				config.VisibilityTimeout = val
			}
		}
		if input.messageRetentionPeriod != "" {
			if val, err := strconv.Atoi(input.messageRetentionPeriod); err == nil {
				config.MessageRetentionPeriod = val
			}
		}
		if input.deliveryDelay != "" {
			if val, err := strconv.Atoi(input.deliveryDelay); err == nil {
				config.DelaySeconds = val
			}
		}
		if input.maximumMessageSize != "" {
			if val, err := strconv.Atoi(input.maximumMessageSize); err == nil {
				config.MaximumMessageSize = val
			}
		}
		if input.receiveMessageWaitTime != "" {
			if val, err := strconv.Atoi(input.receiveMessageWaitTime); err == nil {
				config.ReceiveMessageWaitTime = val
			}
		}

		// FIFO settings
		if config.IsFifo {
			config.ContentBasedDeduplication = input.contentBasedDeduplication
			config.DeduplicationScope = input.deduplicationScope
			config.FifoThroughputLimit = input.fifoThroughputLimit
		}

		// Set loading state and trigger async create
		m.loading = true
		m.loadingMsg = "Creating queue..."

		return m, commands.CreateQueue(m.context, m.client, config)
	}

	// Check if form was aborted
	if m.state.queueCreate.form.State == huh.StateAborted {
		return m.QueueOverviewSwitchPage(msg)
	}

	return m, cmd
}
