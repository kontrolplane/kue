package tui

import (
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
    form  *huh.Form
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    // initialise form on first open
    nameInput := huh.NewInput().Title("Queue name").Value(&m.state.queueCreate.input.name).Validate(func(val string) error {
        if val == "" {
            return fmt.Errorf("name is required")
        }
        return nil
    })

    m.state.queueCreate.form = huh.NewForm(huh.NewGroup(nameInput)).WithTheme(huh.ThemeDracula())

    return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
    if m.state.queueCreate.form == nil {
        return "loading form"
    }
    return m.state.queueCreate.form.View()
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.state.queueCreate.form == nil {
        return m, nil
    }

    var cmd tea.Cmd
    if m.state.queueCreate.form.State() == huh.StateCompleted {
        // user submitted the form
        _, err := kue.CreateQueue(m.client, m.context, kue.CreateQueueInput{Name: m.state.queueCreate.input.name})
        if err != nil {
            m.error = fmt.Sprintf("Error creating queue: %v", err)
        }
        // refresh overview on success
        return m.QueueOverviewSwitchPage(msg)
    }

    m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)
    return m, cmd
}
