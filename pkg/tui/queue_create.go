package tui

import (
    "fmt"
    "log"

    "github.com/charmbracelet/huh"

    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueCreateInput struct {
    name                   string
    visibilityTimeout      int
    messageRetention       int
    deliveryDelay          int
    maximumMessageSize     int
    receiveMessageWaitTime int
}

type queueCreateState struct {
    input queueCreateInput
    form  huh.Form
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    // Reset form each time the page is opened so that previous values don't persist
    m.state.queueCreate.form = buildQueueCreateForm(&m.state.queueCreate.input)
    return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
    if m.state.queueCreate.form == nil {
        m.state.queueCreate.form = buildQueueCreateForm(&m.state.queueCreate.input)
    }
    return m.state.queueCreate.form.View()
}

func (m model) QueueCreateUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    if m.state.queueCreate.form == nil {
        m.state.queueCreate.form = buildQueueCreateForm(&m.state.queueCreate.input)
    }

    m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)

    if m.state.queueCreate.form.State == huh.StateCompleted {
        attr := map[string]string{}
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
            log.Printf("[QueueCreate] Created queue %s (%s)", m.state.queueCreate.input.name, url)
        }

        // Clear input and form for next usage
        m.state.queueCreate.input = queueCreateInput{}
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
