package tui

import (
    "context"
    "log"

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
	return m.SwitchPage(queueCreate), nil
}

func (m model) QueueCreateView() string {
    if m.state.queueCreate.form == nil {
        // initialize form lazily because we need dimensions maybe.
        m.state.queueCreate.form = initQueueCreateForm(&m.state.queueCreate.input)
    }
    return m.state.queueCreate.form.View()
}

type submitCreateQueueMsg struct{}

type cancelCreateQueueMsg struct{}

func (m model) QueueCreateUpdate(msg tea.Msg) (model, tea.Cmd) {

    if m.state.queueCreate.form == nil {
        m.state.queueCreate.form = initQueueCreateForm(&m.state.queueCreate.input)
    }

    // delegate update to form
    var cmd tea.Cmd
    m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)

    switch msg := msg.(type) {
    case submitCreateQueueMsg:
        // call API to create queue
        _, err := kue.CreateQueue(m.client, m.context, kue.QueueCreateInput{
            Name: m.state.queueCreate.input.name,
        })
        if err != nil {
            m.error = err.Error()
        } else {
            // refresh overview page
            log.Printf("[QueueCreate] Queue %s created", m.state.queueCreate.input.name)
        }
        return m.QueueOverviewSwitchPage(msg)
    case cancelCreateQueueMsg:
        return m.QueueOverviewSwitchPage(msg)
    }

    return m, cmd
}

func initQueueCreateForm(input *queueCreateInput) huh.Form {

form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().Title("queue name").Value(&input.name).Placeholder("my-queue"),
        ),
    )

    form.OnSubmit(func(data map[string]any) {
        // bubbletea's huh will send a SubmitMsg automatically; intercept using Update above.
    })

    return form
}
