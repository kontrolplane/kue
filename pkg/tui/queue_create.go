package tui

import (
    "strconv"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/bubbles/key"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueCreateInput holds raw user input (as strings) because huh currently
// returns strings; they are parsed to ints when submitting the form.
// Optional advanced fields live in the Advanced collapsible group. Empty
// strings signal omission.

type queueCreateInput struct {
    name string

    // Advanced section
    visibilityTimeout  string // seconds
    messageRetention   string // seconds
    dlqArn             string
    dlqMaxReceiveCount string
    kmsKeyId           string
}

type queueCreateState struct {
	input queueCreateInput
	form  huh.Form
	openAdvanced bool
}

func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    // initialise form when entering page
    m.state.queueCreate = initQueueCreateState()
    return m.SwitchPage(queueCreate), nil
}

func initQueueCreateState() queueCreateState {
    // Build basic and advanced groups
    st := queueCreateState{}

    basicGroup := []huh.Field{
        huh.NewInput().Title("Queue name").Value(&st.input.name),
    }

    advGroup := huh.NewGroup(
        huh.NewInput().Title("Visibility timeout (s)").Value(&st.input.visibilityTimeout),
        huh.NewInput().Title("Message retention period (s)").Value(&st.input.messageRetention),
        huh.NewInput().Title("DLQ ARN").Value(&st.input.dlqArn),
        huh.NewInput().Title("DLQ max receive count").Value(&st.input.dlqMaxReceiveCount),
        huh.NewInput().Title("KMS Key ID").Value(&st.input.kmsKeyId),
    ).Title("Advanced settings")

    st.form = huh.NewForm(
        huh.NewGroup(basicGroup...),
        advGroup,
    ).WithTheme(huh.ThemeCharm())

    return st
}

func (m model) QueueCreateView() string {
    return m.state.queueCreate.form.View()
}

func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    // handle key toggles
    switch a := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(a, m.keys.Advanced):
            // toggle advanced group collapsed state
            m.state.queueCreate.openAdvanced = !m.state.queueCreate.openAdvanced
            m.state.queueCreate.form.SetCollapsed(!m.state.queueCreate.openAdvanced, 1) // second group
        }
    }

    m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)

    // On submit
    if m.state.queueCreate.form.State == huh.StateCompleted {
        // Convert inputs and call kue.CreateQueue
        qInput := m.state.queueCreate.input
        opts := kue.CreateQueueInput{ Name: qInput.name }
        if v, err := strconv.Atoi(qInput.visibilityTimeout); err == nil && v > 0 {
            opts.VisibilityTimeout = int32(v)
        }
        if v, err := strconv.Atoi(qInput.messageRetention); err == nil && v > 0 {
            opts.MessageRetentionPeriod = int32(v)
        }
        if v, err := strconv.Atoi(qInput.dlqMaxReceiveCount); err == nil && v > 0 {
            opts.DlqMaxReceiveCount = int32(v)
        }
        opts.DlqArn = qInput.dlqArn
        opts.KmsKeyID = qInput.kmsKeyId

        if _, err := kue.CreateQueue(m.client, m.context, opts); err != nil {
            m.error = err.Error()
        }
        // Back to overview
        return m.QueueOverviewSwitchPage(msg)
    }

    return m, cmd
}
