package tui

import (
    "sort"

    "github.com/charmbracelet/huh"
    tea "github.com/charmbracelet/bubbletea"

    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueCreateInput holds the user input of the Create-Queue form. The struct is
// intentionally simple and mirrors one-to-one the input controls we render with
// the huh form library.
//
// A new field – Preset – is added so that the user can pick one of the
// predefined configurations. When a preset is chosen, the advanced fields are
// automatically populated but can still be overridden manually by the user.
//
// NOTE: We store values as strings where the AWS API expects strings to avoid
// premature conversions. The form itself will validate numeric values.

type queueCreateInput struct {
    preset                 string
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

// queueCreateState bundles the current form state together with the huh.Form
// instance. The form pointer is required so that we can call Submit() inside
// the Update loop.

type queueCreateState struct {
    input queueCreateInput
    form  *huh.Form
}

// QueueCreateSwitchPage initialises the Create-Queue form each time we navigate
// to the page so that we have a fresh state.
func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {

    // Build list of preset names (stable order for nicer UX)
    var presetNames []string
    for name := range kue.Presets {
        presetNames = append(presetNames, name)
    }
    sort.Strings(presetNames)

    // Build the huh form with a select for Preset and text inputs for the rest.
    // When a preset is selected, we update the advanced fields inside the form.
    q := &queueCreateState{}

    form := huh.NewForm(
        huh.NewGroup(
            huh.NewSelect[string]().
                Title("Preset").
                Description("Choose a template or leave blank for custom settings").
                Options(stringSliceToOptions(presetNames)...).
                Value(&q.input.preset).
                OnChange(func(s string) {
                    if preset, ok := kue.GetPreset(s); ok {
                        applyPresetToInput(preset, &q.input)
                    }
                }),
            huh.NewInput().Title("Queue name").Value(&q.input.name).Validate(huh.Required[string]()),
            huh.NewSelect[string]().Title("Queue type").Options(
                huh.NewOption("Standard", "standard"),
                huh.NewOption("FIFO", "fifo"),
            ).Value(&q.input.queueType),
            huh.NewInput().Title("Visibility timeout (sec)").Value(&q.input.visibilityTimeout),
            huh.NewInput().Title("Message retention (sec)").Value(&q.input.messageRetention),
            huh.NewInput().Title("Delivery delay (sec)").Value(&q.input.deliveryDelay),
            huh.NewInput().Title("Max message size (bytes)").Value(&q.input.maximumMessageSize),
            huh.NewInput().Title("Receive wait time (sec)").Value(&q.input.receiveMessageWaitTime),
        ),
    ).WithSubmitButton("create queue")

    q.form = form

    m.state.queueCreate = *q

    return m.SwitchPage(queueCreate), form.Init()
}

// QueueCreateView delegates rendering to the huh.Form View.
func (m model) QueueCreateView() string {
    return m.state.queueCreate.form.View()
}

// QueueCreateUpdate passes events to the huh form and handles submission.
func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {

    var cmd tea.Cmd
    var fCmd tea.Cmd

    m.state.queueCreate.form, fCmd = m.state.queueCreate.form.Update(msg)

    // If the form has been submitted we create the queue via the kue package
    if m.state.queueCreate.form.State == huh.StateCompleted {
        attrs := buildAttributesFromInput(m.state.queueCreate.input)
        if _, err := kue.CreateQueue(m.client, m.context, m.state.queueCreate.input.name, attrs); err != nil {
            m.error = err.Error()
        }
        return m.QueueOverviewSwitchPage(msg)
    }

    cmd = fCmd
    return m, cmd
}

// Helper functions ---------------------------------------------------------

func stringSliceToOptions(values []string) []huh.Option[string] {
    var opts []huh.Option[string]
    for _, v := range values {
        opts = append(opts, huh.NewOption(v, v))
    }
    return opts
}

// applyPresetToInput mutates the input struct so that the form fields show the
// preset values to the user.
func applyPresetToInput(preset kue.QueuePreset, in *queueCreateInput) {
    if v, ok := preset["VisibilityTimeout"]; ok {
        in.visibilityTimeout = v
    }
    if v, ok := preset["MessageRetentionPeriod"]; ok {
        in.messageRetention = v
    }
    if v, ok := preset["DelaySeconds"]; ok {
        in.deliveryDelay = v
    }
    if v, ok := preset["MaximumMessageSize"]; ok {
        in.maximumMessageSize = v
    }
    if v, ok := preset["ReceiveMessageWaitTimeSeconds"]; ok {
        in.receiveMessageWaitTime = v
    }
    // When preset is FIFO related we pre-select queue type
    if v, ok := preset["FifoQueue"]; ok && v == "true" {
        in.queueType = "fifo"
    }
}

// buildAttributesFromInput converts the user input into the attribute map that
// the AWS SDK expects.
func buildAttributesFromInput(in queueCreateInput) map[string]string {
    attrs := map[string]string{}

    if in.visibilityTimeout != "" {
        attrs["VisibilityTimeout"] = in.visibilityTimeout
    }
    if in.messageRetention != "" {
        attrs["MessageRetentionPeriod"] = in.messageRetention
    }
    if in.deliveryDelay != "" {
        attrs["DelaySeconds"] = in.deliveryDelay
    }
    if in.maximumMessageSize != "" {
        attrs["MaximumMessageSize"] = in.maximumMessageSize
    }
    if in.receiveMessageWaitTime != "" {
        attrs["ReceiveMessageWaitTimeSeconds"] = in.receiveMessageWaitTime
    }

    // Additional flags for FIFO queues
    if in.queueType == "fifo" {
        attrs["FifoQueue"] = "true"
    }

    return attrs
}
