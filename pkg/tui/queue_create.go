package tui

import (
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/huh"

    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueCreateInput struct {
    name string
}

type queueCreateState struct {
    input queueCreateInput
    form  *huh.Form
}

// QueueCreateSwitchPage initialises the creation form and switches the view.
func (m model) QueueCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    form := huh.NewForm(
        huh.NewGroup(
            huh.NewInput().
                Title("Queue name").
                Value(&m.state.queueCreate.input.name).
                Validate(func(v string) error {
                    if v == "" {
                        return fmt.Errorf("queue name is required")
                    }
                    return nil
                }),
        ),
    ).WithSubmit("Create")

    form.OnSubmit(func() tea.Msg {
        _, err := kue.CreateQueue(m.client, m.context, kue.CreateQueueInput{
            Name: m.state.queueCreate.input.name,
        })
        if err != nil {
            return errorMsg{err}
        }
        // Reset selection and force refresh once we go back.
        m.state.queueOverview.selected = 0
        return tea.KeyMsg{Type: tea.KeyEnter}
    })

    m.state.queueCreate.form = form
    // Reset previous input value
    m.state.queueCreate.input = queueCreateInput{}

    return m.SwitchPage(queueCreate), form.Init()
}

// QueueCreateView renders the form.
func (m model) QueueCreateView() string {
    if m.state.queueCreate.form == nil {
        return "loading…"
    }
    return m.state.queueCreate.form.View()
}

// QueueCreateUpdate delegates update messages to the form.
func (m model) QueueCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    if m.state.queueCreate.form == nil {
        return m, nil
    }
    var cmd tea.Cmd
    *m.state.queueCreate.form, cmd = m.state.queueCreate.form.Update(msg)

    // Handle errors bubbled up from submit.
    if em, ok := msg.(errorMsg); ok {
        m.error = fmt.Sprintf("Error creating queue: %v", em)
    }

    // Once the form completes we switch back to overview (handled via State check)
    if m.state.queueCreate.form.State == huh.StateCompleted {
        return m.QueueOverviewSwitchPage(msg)
    }

    // Allow user to abort with quit
    if km, ok := msg.(tea.KeyMsg); ok {
        if km.Type == tea.KeyEsc || km.String() == "q" {
            return m.QueueOverviewSwitchPage(msg)
        }
    }

    return m, cmd
}

// errorMsg allows bubbling up errors from async commands
// so that they can be handled at a higher level.
type errorMsg struct{ error }
