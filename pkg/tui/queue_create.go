package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueCreateInput struct {
	queueType              string
	name                   string
	region                 string
	visibilityTimeout      string
	messageRetention       string
	deliveryDelay          string
	maximumMessageSize     string
	receiveMessageWaitTime string
}

// Helper to create the Huh form
func initQueueCreateForm(m *model) huh.Form {
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Queue name").
				Value(&m.state.queueCreate.input.name).
				Validate(func(v string) error {
					if v == "" {
						return fmt.Errorf("name cannot be empty")
					}
					return nil
				}),
			huh.NewInput().
				Title("Visibility timeout (seconds, optional)").
				Value(&m.state.queueCreate.input.visibilityTimeoutType),
			huh.NewInput().
				Title("Message retention (seconds, optional)").
				Value(&m.state.queueCreate.input.messageRetentionType),
			huh.NewInput().
				Title("Delivery delay (seconds, optional)").
				Value(&m.state.queueCreate.input.deliveryDelayType),
			huh.NewInput().
				Title("Maximum message size (bytes, optional)").
				Value(&m.state.queueCreate.input.maximumMessageSize),
			huh.NewInput().
				Title("Receive message wait time (seconds, optional)").
				Value(&m.state.queueCreate.input.receiveMessageWaitTime),
		),
	).WithTheme(huh.ThemeCatppuccin())

	return f
}
type queueCreateState struct {
	input queueCreateInput
	form  huh.Form
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
	// Render Huh form if it exists
	if m.state.queueCreate.form != nil {
		return m.state.queueCreate.form.View()
	}
	return "Creating a new queue..."
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Initialize the form the first time we land on the page
	if m.state.queueCreate.form == nil {
		m.state.queueCreate.form = initQueueCreateForm(&m)
	}

	// Let the form handle the message
	if m.state.queueCreate.form != nil {
		m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)

		if m.state.queueCreate.form.Succeeded() {
			// Gather values and call AWS
			attrs := map[string]string{
				"VisibilityTimeout":               m.state.queueCreate.input.visibilityTimeout,
				"MessageRetentionPeriod":          m.state.queueCreate.input.messageRetention,
				"DelaySeconds":                    m.state.queueCreate.input.deliveryDelay,
				"MaximumMessageSize":              m.state.queueCreate.input.maximumMessageSize,
				"ReceiveMessageWaitTimeSeconds":   m.state.queueCreate.input.receiveMessageWaitTime,
			}
			if err := kue.CreateQueue(m.client, m.context, m.state.queueCreate.input.name, attrs);
			err != nil {
				m.error = fmt.Sprintf("Error creating queue: %v", err)
			} else {
				// refresh overview
				return m.QueueOverviewSwitchPage(msg)
			}
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit) {
			return m.QueueOverviewSwitchPage(msg)
		}
	}

	return m, cmd
}
