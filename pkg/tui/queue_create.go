package tui

import (
//	"context" // no longer used
	"fmt"
	"strings"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"

	kue "github.com/kontrolplane/kue/pkg/kue"
)
type queueCreateInput struct {
	queueType              string
	name                   string
	region                 string
	visibilityTimeout      string
	visibilityTimeoutType  string
	messageRetention       string
	messageRetentionType   string
	deliveryDelay          string
	deliveryDelayType      string
	maximumMessageSize     string
	receiveMessageWaitTime string
}

// validation function for queue name
func validateQueueName(name string, isFIFO bool) error {
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("queue name cannot be empty")
	}
	if len(name) > 80 {
		return fmt.Errorf("queue name cannot exceed 80 characters")
	}
	if isFIFO && !strings.HasSuffix(name, ".fifo") {
		return fmt.Errorf("FIFO queue name must end with .fifo")
	}
	return nil
}

type queueCreateState struct {
	input queueCreateInput
	form  *huh.Form
	focus bool
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	// reset form each time we enter
	m.state.queueCreate.focus = true
	cmd := m.initQueueCreateForm()
	return m.SwitchPage(queueCreate), cmd
}

// initialize the form and return initial command
func (m *model) initQueueCreateForm() tea.Cmd {
	// default values
	m.state.queueCreate.input.queueType = "Standard"
	m.state.queueCreate.input.visibilityTimeoutType = "seconds"
	m.state.queueCreate.input.messageRetentionType = "seconds"
	m.state.queueCreate.input.deliveryDelayType = "seconds"

	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Queue Type").
			Options(
				huh.NewOption("Standard", "Standard"),
				huh.NewOption("FIFO", "FIFO"),
			).
			Value(&m.state.queueCreate.input.queueType),
		huh.NewInput().
			Title("Queue Name").
			Value(&m.state.queueCreate.input.name).
			Validate(func(v string) error {
				isFIFO := m.state.queueCreate.input.queueType == "FIFO"
				return validateQueueName(v, isFIFO)
			}),
		huh.NewInput().
			Title("Region (leave blank for default)").
			Value(&m.state.queueCreate.input.region),
		huh.NewInput().
			Title("Visibility Timeout (seconds)").
			Value(&m.state.queueCreate.input.visibilityTimeout).
			Validate(huh.ValidateInt),
		huh.NewInput().
			Title("Message Retention (seconds)").
			Value(&m.state.queueCreate.input.messageRetention).
			Validate(huh.ValidateInt),
		huh.NewInput().
			Title("Delivery Delay (seconds)").
			Value(&m.state.queueCreate.input.deliveryDelay).
			Validate(huh.ValidateInt),
		huh.NewInput().
			Title("Maximum Message Size (bytes)").
			Value(&m.state.queueCreate.input.maximumMessageSize).
			Validate(huh.ValidateInt),
		huh.NewInput().
			Title("Receive Message Wait Time (seconds)").
			Value(&m.state.queueCreate.input.receiveMessageWaitTime).
			Validate(huh.ValidateInt),
	)).WithTheme(huh.ThemeCatppuccin())

	m.state.queueCreate.form = form

	return form.Init()
}

func (m model) QueueCreateView() string {
	if m.state.queueCreate.form == nil {
		return "Initializing form..."
	}
	return m.state.queueCreate.form.View()
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.state.queueCreate.form == nil {
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m.QueueOverviewSwitchPage(msg)
		}
	}

	var fcmd tea.Cmd
	m.state.queueCreate.form, fcmd = m.state.queueCreate.form.Update(msg)
	cmd = fcmd

	if m.state.queueCreate.form.Submitted() {
		// call AWS to create queue
		queueType := m.state.queueCreate.input.queueType
	// convert string fields to int
	toInt := func(s string) int {
		v, _ := strconv.Atoi(strings.TrimSpace(s))
		return v
	}
	attrs := kue.CreateQueueAttributes{
		VisibilityTimeout:      toInt(m.state.queueCreate.input.visibilityTimeout),
		MessageRetentionPeriod: toInt(m.state.queueCreate.input.messageRetention),
		DelaySeconds:           toInt(m.state.queueCreate.input.deliveryDelay),
		MaximumMessageSize:     toInt(m.state.queueCreate.input.maximumMessageSize),
		ReceiveMessageWaitTime: toInt(m.state.queueCreate.input.receiveMessageWaitTime),
	}

		isFIFO := queueType == "FIFO"
		if isFIFO {
			attrs.FifoQueue = true
		}

		if err := kue.CreateQueue(m.client, m.context, m.state.queueCreate.input.name, attrs); err != nil {
			m.error = fmt.Sprintf("Error creating queue: %v", err)
		}

		return m.QueueOverviewSwitchPage(msg)
	}

	return m, cmd
}
