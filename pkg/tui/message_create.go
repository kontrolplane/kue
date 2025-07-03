package tui

import (
    "encoding/json"
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/textarea"

    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueMessageCreateState holds state for the compose modal.
// It embeds a bubbles textarea component and tracks whether the modal is
// expanded as well as any transient error returned from JSON formatting.
//
// The selected queue to which the message will be sent is stored in
// m.state.queueDetails.queue while this page is active.

type queueMessageCreateState struct {
    textarea textarea.Model
    expanded bool
    errMsg   string
}

// QueueMessageCreateSwitchPage initialises a new compose modal and switches the page.
func (m model) QueueMessageCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    ta := textarea.New()
    ta.Placeholder = "Enter message body (JSON or plain text)"
    ta.SetWidth(80)
    ta.SetHeight(10)
    ta.ShowLineNumbers = false
    ta.Focus()

    m.state.queueMessageCreate = queueMessageCreateState{
        textarea: ta,
        expanded: false,
    }

    return m.SwitchPage(queueMessageCreate), nil
}

func (m model) QueueMessageCreateView() string {
    // resize textarea if expanded/fullscreen
    if m.state.queueMessageCreate.expanded {
        ta := m.state.queueMessageCreate.textarea
        ta.SetWidth(max(10, m.width-4))
        ta.SetHeight(max(5, m.height-6))
        m.state.queueMessageCreate.textarea = ta
    } else {
        ta := m.state.queueMessageCreate.textarea
        ta.SetWidth(80)
        ta.SetHeight(10)
        m.state.queueMessageCreate.textarea = ta
    }

    view := m.state.queueMessageCreate.textarea.View()

    if m.state.queueMessageCreate.errMsg != "" {
        view += fmt.Sprintf("\n\n[error] %s", m.state.queueMessageCreate.errMsg)
    }

    help := "ctrl+e expand | ctrl+f format json | ctrl+s send | esc cancel"

    return view + "\n" + help
}

func (m model) QueueMessageCreateUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit): // esc/q
            return m.QueueDetailsSwitchPage(msg)
        case msg.String() == "ctrl+e":
            m.state.queueMessageCreate.expanded = !m.state.queueMessageCreate.expanded
        case msg.String() == "ctrl+f":
            formatted, err := FormatJSON(m.state.queueMessageCreate.textarea.Value())
            if err != nil {
                m.state.queueMessageCreate.errMsg = err.Error()
            } else {
                m.state.queueMessageCreate.textarea.SetValue(formatted)
                m.state.queueMessageCreate.errMsg = ""
            }
        case msg.String() == "ctrl+s":
            body := m.state.queueMessageCreate.textarea.Value()
            _, err := kue.SendMessage(m.client, m.context, m.state.queueDetails.queue.Url, body, 0)
            if err != nil {
                m.state.queueMessageCreate.errMsg = err.Error()
                return m, nil
            }
            // After successful send, switch back to details page (refreshing)
            return m.QueueDetailsSwitchPage(msg)
        }
    }

    m.state.queueMessageCreate.textarea, cmd = m.state.queueMessageCreate.textarea.Update(msg)
    return m, cmd
}

// FormatJSON attempts to prettify JSON input.
// It returns the pretty version on success, otherwise original string with an error.
func FormatJSON(input string) (string, error) {
    var i interface{}
    if err := json.Unmarshal([]byte(input), &i); err != nil {
        return input, fmt.Errorf("invalid JSON: %w", err)
    }
    b, err := json.MarshalIndent(i, "", "  ")
    if err != nil {
        return input, err
    }
    return string(b), nil
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
