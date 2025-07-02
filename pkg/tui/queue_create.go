package tui

import (
    "log"
    "fmt"

    "github.com/charmbracelet/huh"

    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueCreateInput struct {
	queueType              string
	name                   string
	region                 string
	visibilityTimeout      int
	visibilityTimeoutType  string
	messageRetention       int
	messageRetentionType   string
	deliveryDelay          int
	deliveryDelayType      string
	maximumMessageSize     int
	receiveMessageWaitTime int
}

type queueCreateState struct {
	input queueCreateInput
	form  huh.Form
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
    // render form
    if m.state.queueCreate.form == nil {
        m.state.queueCreate.form = buildQueueCreateForm(&m.state.queueCreate.input)
    }

    return m.state.queueCreate.form.View()
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    if m.state.queueCreate.form == nil {
        m.state.queueCreate.form = buildQueueCreateForm(&m.state.queueCreate.input)
    }

    m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)

    // when submitted, create queue and go back to overview
    if m.state.queueCreate.form.State == huh.StateCompleted {
        // call create queue
        attr := map[string]string{}
        // convert ints to strings where non zero
        if m.state.queueCreate.input.visibilityTimeout > 0 {
            attr["VisibilityTimeout"] = fmt.Sprintf("%d", m.state.queueCreate.input.visibilityTimeout)
        }
        if m.state.queueCreate.input.messageRetention > 0 {
            attr["MessageRetentionPeriod"] = fmt.Sprintf("%d", m.state.queueCreate.input.messageRetention)
        }
        if m.state.queueCreate.input.deliveryDelay > 0 {
            attr["DelaySeconds"] = fmt.Sprintf("%d", m.state.queueCreate.input.deliveryDelay)
        }
        if m.state.queueCreate.input.maximumMessageSize > 0 {
            attr["MaximumMessageSize"] = fmt.Sprintf("%d", m.state.queueCreate.input.maximumMessageSize)
        }
        if m.state.queueCreate.input.receiveMessageWaitTime > 0 {
            attr["ReceiveMessageWaitTimeSeconds"] = fmt.Sprintf("%d", m.state.queueCreate.input.receiveMessageWaitTime)
        }

        url, err := kue.CreateQueue(m.client, m.context, m.state.queueCreate.input.name, attr)
        if err != nil {
            m.error = fmt.Sprintf("Error creating queue: %v", err)
        } else {
            log.Printf("[QueueCreate] created queue %s (%s)", m.state.queueCreate.input.name, url)
        }
        // reset form
        m.state.queueCreate.form = nil
        return m.QueueOverviewSwitchPage(msg)
    }

    return m, cmd
}

func buildQueueCreateForm(input *queueCreateInput) huh.Form {
    return huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Queue Name").
                Value(&input.name).
                Validate(func(v string) error {
                    if v == "" {
                        return fmt.Errorf("queue name required")
                    }
                    return nil
                }),
            huh.NewNumberInput().
                Title("Visibility Timeout (seconds)").
                Value(&input.visibilityTimeout),
            huh.NewNumberInput().
                Title("Message Retention (seconds)").
                Value(&input.messageRetention),
            huh.NewNumberInput().
                Title("Delivery Delay (seconds)").
                Value(&input.deliveryDelay),
            huh.NewNumberInput().
                Title("Maximum Message Size (bytes)").
                Value(&input.maximumMessageSize),
            huh.NewNumberInput().
                Title("Receive Message WaitTime (seconds)").
                Value(&input.receiveMessageWaitTime),
        ),
    )
}
