package tui

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"

	"github.com/kontrolplane/kue/pkg/kue"
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

// dirty reports whether anything was typed that leaving the form would lose.
func (in *queueCreateInput) dirty() bool {
	for _, v := range []string{in.name, in.visibilityTimeout, in.messageRetentionPeriod, in.deliveryDelay, in.maximumMessageSize, in.receiveMessageWaitTime} {
		if strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}

type queueCreateState struct {
	input       *queueCreateInput
	form        *huh.Form
	stepOf      map[huh.Field]int
	currentStep int
}

// formWidth leaves the form room on both sides, like the other pages, up to 100 columns.
func formWidth() int { return min(100, contentWidth-8) }

// formHeight is the height the form gets below its step header.
func formHeight() int { return contentHeight - 4 }

var formSteps = []string{"basic", "messages", "advanced", "fifo"}

// newQueueCreateForm builds the form, and maps each field to the index of its step in formSteps.
// The fifo step only shows for a fifo queue.
func newQueueCreateForm(input *queueCreateInput) (*huh.Form, map[huh.Field]int) {
	steps := [][]huh.Field{
		{
			huh.NewInput().
				Title("queue name").
				Description("letters, digits, hyphens and underscores, at most 80 characters").
				Placeholder("orders").
				Value(&input.name).
				Validate(validateQueueName),

			huh.NewSelect[string]().
				Title("queue type").
				Description("standard: best-effort ordering, higher throughput. fifo: guaranteed ordering").
				Options(
					huh.NewOption("standard", "standard"),
					huh.NewOption("fifo", "fifo"),
				).
				Value(&input.queueType),
		},
		{
			huh.NewInput().
				Title("visibility timeout").
				Description("seconds a message is hidden after being received (0-43200)").
				Placeholder("30").
				Value(&input.visibilityTimeout).
				Validate(validateIntRange(0, 43200)),

			huh.NewInput().
				Title("message retention period").
				Description("seconds messages are kept before deletion (60-1209600)").
				Placeholder("345600").
				Value(&input.messageRetentionPeriod).
				Validate(validateIntRange(60, 1209600)),

			huh.NewInput().
				Title("delivery delay").
				Description("seconds before messages become visible (0-900)").
				Placeholder("0").
				Value(&input.deliveryDelay).
				Validate(validateIntRange(0, 900)),
		},
		{
			huh.NewInput().
				Title("maximum message size").
				Description("maximum message size in bytes (1024-262144)").
				Placeholder("262144").
				Value(&input.maximumMessageSize).
				Validate(validateIntRange(1024, 262144)),

			huh.NewInput().
				Title("receive wait time").
				Description("long polling wait time in seconds (0-20)").
				Placeholder("0").
				Value(&input.receiveMessageWaitTime).
				Validate(validateIntRange(0, 20)),
		},
		{
			huh.NewConfirm().
				Title("content based deduplication").
				Description("deduplicate messages based on a hash of their body").
				Value(&input.contentBasedDeduplication),

			huh.NewSelect[string]().
				Title("deduplication scope").
				Description("where message deduplication applies").
				Options(
					huh.NewOption("queue", "queue"),
					huh.NewOption("message group", "messageGroup"),
				).
				Value(&input.deduplicationScope),

			huh.NewSelect[string]().
				Title("throughput limit").
				Description("how the throughput quota is allocated").
				Options(
					huh.NewOption("per queue", "perQueue"),
					huh.NewOption("per message group id", "perMessageGroupId"),
				).
				Value(&input.fifoThroughputLimit),
		},
	}

	stepOf := map[huh.Field]int{}
	groups := make([]*huh.Group, len(steps))
	for i, fields := range steps {
		for _, f := range fields {
			stepOf[f] = i
		}
		groups[i] = huh.NewGroup(fields...).Title(formSteps[i])
	}
	groups[3] = groups[3].WithHideFunc(func() bool { return input.queueType != "fifo" })

	form := huh.NewForm(groups...).
		WithTheme(styles.FormTheme()).
		WithShowHelp(false).
		WithWidth(formWidth()).
		WithHeight(formHeight()).
		WithShowErrors(false)
	return form, stepOf
}

func validateQueueName(s string) error {
	s = strings.TrimSpace(s)
	switch {
	case s == "":
		return fmt.Errorf("queue name is required")
	case len(s) > 80:
		return fmt.Errorf("queue name must be 80 characters or less")
	case !queueNameRegex.MatchString(s):
		return fmt.Errorf("only alphanumeric characters, hyphens, and underscores allowed")
	}
	return nil
}

func validateIntRange(min, max int) func(string) error {
	return func(s string) error {
		if s = strings.TrimSpace(s); s == "" {
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

func (m model) QueueCreateSwitchPage() (model, tea.Cmd) {
	m.error = ""
	m.armed = false
	m.state.queueCreate.input = &queueCreateInput{
		queueType:           "standard",
		deduplicationScope:  "queue",
		fifoThroughputLimit: "perQueue",
	}
	return m.openQueueCreateForm()
}

// openQueueCreateForm shows a new form on the current input, which keeps what was entered.
func (m model) openQueueCreateForm() (model, tea.Cmd) {
	c := &m.state.queueCreate
	c.form, c.stepOf = newQueueCreateForm(c.input)
	c.currentStep = 0
	return m.SwitchPage(queueCreate), c.form.Init()
}

// reopenQueueCreate returns to the form after the queue could not be created, keeping the input.
func (m model) reopenQueueCreate(reason string) (model, tea.Cmd) {
	m, cmd := m.openQueueCreateForm()
	m.error = reason
	return m, cmd
}

func (m model) QueueCreateView() string {
	if m.state.queueCreate.form == nil {
		return "loading…"
	}
	content := lipgloss.JoinVertical(lipgloss.Left,
		m.renderFormHeader(),
		m.state.queueCreate.form.View(),
	)
	return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Top, content)
}

// renderFormHeader renders the progress indicator showing the current form step.
func (m model) renderFormHeader() string {
	c := m.state.queueCreate
	names := formSteps
	if c.input == nil || c.input.queueType != "fifo" {
		names = formSteps[:3]
	}
	var steps []string
	for i, step := range names {
		switch {
		case i < c.currentStep:
			steps = append(steps, styles.Render(styles.S("✓ ", styles.ToneAccent), styles.S(step, styles.ToneMuted)))
		case i == c.currentStep:
			steps = append(steps, styles.Render(styles.B("● ", styles.ToneAccent), styles.B(step, styles.ToneText)))
		default:
			steps = append(steps, styles.Render(styles.S("○ ", styles.ToneFaint), styles.S(step, styles.ToneFaint)))
		}
	}
	line := strings.Join(steps, styles.Faint("   ───   "))
	return lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.PlaceHorizontal(formWidth(), lipgloss.Center, line),
		lipgloss.NewStyle().Foreground(styles.P.Rule).Render(strings.Repeat("─", formWidth())),
		lipgloss.PlaceHorizontal(formWidth(), lipgloss.Left, m.renderFormError()),
	)
}

// renderFormError shows why the form does not move on. huh adds its errors below a group sized
// without them, where they are cut off, so the line the header keeps free shows them instead.
func (m model) renderFormError() string {
	f := m.state.queueCreate.form
	if f == nil {
		return ""
	}
	errs := f.Errors()
	if len(errs) == 0 {
		return ""
	}
	return styles.Render(styles.S("✗ ", styles.ToneDanger), styles.S(truncate(errs[0].Error(), formWidth()-2), styles.ToneDanger))
}

func (m model) QueueCreateUpdate(msg tea.Msg) (model, tea.Cmd) {
	c := &m.state.queueCreate
	if c.form == nil || m.loading {
		return m, nil
	}

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, m.keys.Back) {
		if c.input.dirty() && !m.armed {
			m.armed = true
			return m.setStatus(discardPrompt, styles.ToneWarning)
		}
		return m.QueueOverviewGoBack()
	}

	form, cmd := c.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		c.form = f
		c.currentStep = c.stepOf[f.GetFocusedField()]
	}

	switch c.form.State {
	case huh.StateCompleted:
		return m.submitQueueCreate()
	case huh.StateAborted:
		return m.QueueOverviewGoBack()
	}

	return m, cmd
}

// buildQueueConfig converts the form input into a queue configuration.
func buildQueueConfig(input *queueCreateInput) (kue.QueueConfig, error) {
	name := strings.TrimSpace(input.name)
	if err := validateQueueName(name); err != nil {
		return kue.QueueConfig{}, err
	}

	config := kue.QueueConfig{
		Name:   name,
		IsFifo: input.queueType == "fifo",
	}
	for _, f := range []struct {
		value string
		dst   *int
	}{
		{input.visibilityTimeout, &config.VisibilityTimeout},
		{input.messageRetentionPeriod, &config.MessageRetentionPeriod},
		{input.deliveryDelay, &config.DelaySeconds},
		{input.maximumMessageSize, &config.MaximumMessageSize},
		{input.receiveMessageWaitTime, &config.ReceiveMessageWaitTime},
	} {
		if val, err := strconv.Atoi(strings.TrimSpace(f.value)); err == nil {
			*f.dst = val
		}
	}

	if config.IsFifo {
		config.ContentBasedDeduplication = input.contentBasedDeduplication
		config.DeduplicationScope = input.deduplicationScope
		config.FifoThroughputLimit = input.fifoThroughputLimit
	}
	return config, nil
}

// submitQueueCreate validates input and triggers async queue creation.
func (m model) submitQueueCreate() (model, tea.Cmd) {
	config, err := buildQueueConfig(m.state.queueCreate.input)
	if err != nil {
		return m.reopenQueueCreate(err.Error())
	}
	m.busy, m.loading, m.loadingMsg = true, true, "creating queue…"
	return m, commands.CreateQueue(m.context, m.client, config)
}
