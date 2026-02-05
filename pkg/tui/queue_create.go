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

// Queue name validation: alphanumeric, hyphens, and underscores only.
var queueNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// queueCreateInput holds form field values during queue creation.
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
	input       *queueCreateInput
	form        *huh.Form
	currentStep int
}

const formWidth = 100

// newQueueCreateForm builds the multi-step queue creation form.
// Steps: Basic → Messages → Advanced → FIFO (conditional).
func newQueueCreateForm(input *queueCreateInput) *huh.Form {
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Queue Name").
				Description("Alphanumeric characters, hyphens, and underscores only (1-80 chars)").
				Placeholder("my-queue-name").
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
				Description("Standard: best-effort ordering, higher throughput. FIFO: guaranteed ordering.").
				Options(
					huh.NewOption("Standard", "standard"),
					huh.NewOption("FIFO", "fifo"),
				).
				Value(&input.queueType),
		).Title("Basic Configuration").
			Description("Required settings for creating a new queue"),

		huh.NewGroup(
			huh.NewInput().
				Title("Visibility Timeout").
				Description("Seconds a message is hidden after being received (0-43200)").
				Placeholder("30").
				Value(&input.visibilityTimeout).
				Validate(validateIntRange(0, 43200)),

			huh.NewInput().
				Title("Message Retention Period").
				Description("Seconds messages are kept before deletion (60-1209600)").
				Placeholder("345600").
				Value(&input.messageRetentionPeriod).
				Validate(validateIntRangeOrEmpty(60, 1209600)),

			huh.NewInput().
				Title("Delivery Delay").
				Description("Seconds before messages become visible (0-900)").
				Placeholder("0").
				Value(&input.deliveryDelay).
				Validate(validateIntRange(0, 900)),
		).Title("Message Settings").
			Description("Configure message visibility and retention"),

		huh.NewGroup(
			huh.NewInput().
				Title("Maximum Message Size").
				Description("Maximum message size in bytes (1024-262144)").
				Placeholder("262144").
				Value(&input.maximumMessageSize).
				Validate(validateIntRangeOrEmpty(1024, 262144)),

			huh.NewInput().
				Title("Receive Wait Time").
				Description("Long polling wait time in seconds (0-20)").
				Placeholder("0").
				Value(&input.receiveMessageWaitTime).
				Validate(validateIntRange(0, 20)),
		).Title("Advanced Settings").
			Description("Fine-tune queue behavior"),

		huh.NewGroup(
			huh.NewConfirm().
				Title("Content-Based Deduplication").
				Description("Automatically deduplicate messages based on body content").
				Value(&input.contentBasedDeduplication),

			huh.NewSelect[string]().
				Title("Deduplication Scope").
				Description("Scope for message deduplication").
				Options(
					huh.NewOption("Queue", "queue"),
					huh.NewOption("Message Group", "messageGroup"),
				).
				Value(&input.deduplicationScope),

			huh.NewSelect[string]().
				Title("Throughput Limit").
				Description("Throughput quota allocation").
				Options(
					huh.NewOption("Per Queue", "perQueue"),
					huh.NewOption("Per Message Group ID", "perMessageGroupId"),
				).
				Value(&input.fifoThroughputLimit),
		).Title("FIFO Settings").
			Description("Configure FIFO-specific queue behavior").
			WithHideFunc(func() bool { return input.queueType != "fifo" }),
	).
		WithTheme(styles.FormTheme()).
		WithShowHelp(true).
		WithWidth(formWidth).
		WithShowErrors(true)

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
	m.error = ""
	m.state.queueCreate.input = &queueCreateInput{
		queueType:           "standard",
		deduplicationScope:  "queue",
		fifoThroughputLimit: "perQueue",
	}
	m.state.queueCreate.form = newQueueCreateForm(m.state.queueCreate.input)
	m.state.queueCreate.currentStep = 0
	return m.SwitchPage(queueCreate), m.state.queueCreate.form.Init()
}

func (m model) QueueCreateView() string {
	if m.state.queueCreate.form == nil {
		return "Loading..."
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		m.renderFormHeader(),
		m.state.queueCreate.form.View(),
	)
	return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Top, content)
}

// detectFormStep determines the current step by checking for unique field titles.
func detectFormStep(view string) int {
	switch {
	case strings.Contains(view, "Content-Based Deduplication"),
		strings.Contains(view, "Deduplication Scope"),
		strings.Contains(view, "Throughput Limit"):
		return 3 // FIFO
	case strings.Contains(view, "Maximum Message Size"),
		strings.Contains(view, "Receive Wait Time"):
		return 2 // Advanced
	case strings.Contains(view, "Visibility Timeout"),
		strings.Contains(view, "Message Retention"),
		strings.Contains(view, "Delivery Delay"):
		return 1 // Messages
	default:
		return 0 // Basic
	}
}

// renderFormHeader renders the progress indicator showing current form step.
func (m model) renderFormHeader() string {
	isFifo := m.state.queueCreate.input != nil && m.state.queueCreate.input.queueType == "fifo"

	steps := []string{"1. Basic", "2. Messages", "3. Advanced"}
	if isFifo {
		steps = append(steps, "4. FIFO")
	}

	currentStep := m.state.queueCreate.currentStep
	if !isFifo && currentStep > 2 {
		currentStep = 2
	}

	var stepViews []string
	for i, step := range steps {
		style := lipgloss.NewStyle().PaddingRight(3)
		switch {
		case i < currentStep:
			style = style.Foreground(styles.AccentColor)
		case i == currentStep:
			style = style.Foreground(styles.TextLight).Bold(true)
		default:
			style = style.Foreground(styles.DarkGray)
		}
		stepViews = append(stepViews, style.Render(step))
	}

	return lipgloss.NewStyle().
		Width(formWidth).
		Align(lipgloss.Center).
		PaddingBottom(1).
		MarginBottom(1).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(styles.BorderColor).
		Render(lipgloss.JoinHorizontal(lipgloss.Center, stepViews...))
}

func (m model) QueueCreateUpdate(msg tea.Msg) (model, tea.Cmd) {
	if m.state.queueCreate.form == nil {
		return m, nil
	}

	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "esc" {
		return m.QueueOverviewSwitchPage(msg)
	}

	form, cmd := m.state.queueCreate.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.state.queueCreate.form = f
		m.state.queueCreate.currentStep = detectFormStep(f.View())
	}

	switch m.state.queueCreate.form.State {
	case huh.StateCompleted:
		return m.submitQueueCreate(msg)
	case huh.StateAborted:
		return m.QueueOverviewSwitchPage(msg)
	}

	return m, cmd
}

// submitQueueCreate validates input and triggers async queue creation.
func (m model) submitQueueCreate(msg tea.Msg) (model, tea.Cmd) {
	input := m.state.queueCreate.input
	queueName := strings.TrimSpace(input.name)

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

	// Parse optional numeric settings
	if val, err := strconv.Atoi(input.visibilityTimeout); err == nil {
		config.VisibilityTimeout = val
	}
	if val, err := strconv.Atoi(input.messageRetentionPeriod); err == nil {
		config.MessageRetentionPeriod = val
	}
	if val, err := strconv.Atoi(input.deliveryDelay); err == nil {
		config.DelaySeconds = val
	}
	if val, err := strconv.Atoi(input.maximumMessageSize); err == nil {
		config.MaximumMessageSize = val
	}
	if val, err := strconv.Atoi(input.receiveMessageWaitTime); err == nil {
		config.ReceiveMessageWaitTime = val
	}

	if config.IsFifo {
		config.ContentBasedDeduplication = input.contentBasedDeduplication
		config.DeduplicationScope = input.deduplicationScope
		config.FifoThroughputLimit = input.fifoThroughputLimit
	}

	m.loading = true
	m.loadingMsg = "Creating queue..."
	return m, commands.CreateQueue(m.context, m.client, config)
}
