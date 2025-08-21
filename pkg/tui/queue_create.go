package tui

import (
    "fmt"
    "strconv"

    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/huh"

    tea "github.com/charmbracelet/bubbletea"

    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueCreateInput holds raw input values from the form.
// All optional numeric values are captured as string first and validated later.
// This avoids partial state mutation issues while user is typing.
type queueCreateInput struct {
    name                   string
    visibilityTimeout      string
    messageRetention       string
    deliveryDelay          string
    maximumMessageSize     string
    receiveMessageWaitTime string
}

type queueCreateState struct {
    input queueCreateInput
    form  huh.Form
}

// initQueueCreateForm creates and returns a huh.Form bound to the provided state pointer.
func initQueueCreateForm(s *queueCreateState) huh.Form {
    f := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("queue name (required)").
                Value(&s.input.name).
                Validate(func(v string) error {
                    if v == "" {
                        return fmt.Errorf("queue name is required")
                    }
                    return nil
                }),
            huh.NewInput().
                Title("visibility timeout (seconds, optional)").
                Value(&s.input.visibilityTimeout),
            huh.NewInput().
                Title("message retention period (seconds, optional)").
                Value(&s.input.messageRetention),
            huh.NewInput().
                Title("delivery delay (seconds, optional)").
                Value(&s.input.deliveryDelay),
            huh.NewInput().
                Title("maximum message size (bytes, optional)").
                Value(&s.input.maximumMessageSize),
            huh.NewInput().
                Title("receive message wait time (seconds, optional)").
                Value(&s.input.receiveMessageWaitTime),
        ),
    )
    return *f
}

// QueueCreateSwitchPage prepares the form and switches to the queueCreate page.
func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    m.state.queueCreate = queueCreateState{}
    form := initQueueCreateForm(&m.state.queueCreate)
    m.state.queueCreate.form = form
    return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
    return m.state.queueCreate.form.View()
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    // First let the form process the message
    m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)

    switch kt := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(kt, m.keys.Quit): // Abort create – return to overview
            return m.QueueOverviewSwitchPage(msg)
        }
    }

    // When the form reaches the completed state try to create the queue and return
    if m.state.queueCreate.form.State == huh.StateCompleted {
        attrs := make(map[string]string)

        // helper to set integer attributes if user provided value
        parseAndSet := func(val string, attr string) {
            if val == "" {
                return
            }
            if _, err := strconv.Atoi(val); err == nil {
                attrs[attr] = val
            }
        }

        parseAndSet(m.state.queueCreate.input.visibilityTimeout, "VisibilityTimeout")
        parseAndSet(m.state.queueCreate.input.messageRetention, "MessageRetentionPeriod")
        parseAndSet(m.state.queueCreate.input.deliveryDelay, "DelaySeconds")
        parseAndSet(m.state.queueCreate.input.maximumMessageSize, "MaximumMessageSize")
        parseAndSet(m.state.queueCreate.input.receiveMessageWaitTime, "ReceiveMessageWaitTimeSeconds")

        if err := kue.CreateQueue(m.client, m.context, m.state.queueCreate.input.name, attrs); err != nil {
            m.error = fmt.Sprintf("Error creating queue: %v", err)
        } else {
            m.error = fmt.Sprintf("Successfully created queue: %s", m.state.queueCreate.input.name)
        }

        // After creating (or failure) return to overview which will reload list when switching page
        return m.QueueOverviewSwitchPage(msg)
    }

    return m, cmd
}
