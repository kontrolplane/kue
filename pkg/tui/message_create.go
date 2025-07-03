package tui

import (
    "bytes"
    "encoding/json"
    "log"

    "github.com/charmbracelet/bubbles/textarea"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/lipgloss"
)

type queueMessageCreateInput struct {
	messageBody   string
	deliveryDelay string
}

type queueMessageCreateState struct {
	input queueMessageCreateInput
    form  huh.Form
    fullScreen bool
    textarea   textarea.Model
}

// key toggles for view
// Use key bindings defined in keys package

func (m model) QueueMessageCreateSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    // initialize textarea when switching to this page
    ta := textarea.New()
    ta.Placeholder = "Enter message body (JSON)..."
    ta.Focus()
    ta.CharLimit = 0 // unlimited

    m.state.queueMessageCreate = queueMessageCreateState{
        textarea: ta,
    }

    return m.SwitchPage(queueMessageCreate), nil
}

func (m model) QueueMessageCreateView() string {
    s := m.state.queueMessageCreate
    var header string = "Compose Message (" + m.keys.Fullscreen.Help().Key + " fullscreen, " + m.keys.Format.Help().Key + " format JSON, esc to back)"

    if s.fullScreen {
        return header + "\n" + s.textarea.View()
    }

    // Render minimal modal style (fixed width/height centre of screen)
    width := 80
    height := 10
    modalStyle := lipgloss.NewStyle().Width(width).Height(height + 2) // +2 for header
    content := header + "\n" + lipgloss.NewStyle().Width(width).Height(height).Render(s.textarea.View())
    return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalStyle.Render(content))
}

func (m model) QueueMessageCreateUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "esc":
            // go back to previous page
            return m.SwitchPage(m.previous), nil
        case "ctrl+f":
            m.state.queueMessageCreate.fullScreen = !m.state.queueMessageCreate.fullScreen
        case "ctrl+j":
            // attempt to pretty-format JSON
            body := m.state.queueMessageCreate.textarea.Value()
            var pretty bytes.Buffer
            if err := json.Indent(&pretty, []byte(body), "", "  "); err == nil {
                m.state.queueMessageCreate.textarea.SetValue(pretty.String())
            } else {
                log.Println("[QueueMessageCreate] invalid json:", err)
            }
        }
    }

    m.state.queueMessageCreate.textarea, cmd = m.state.queueMessageCreate.textarea.Update(msg)
    return m, cmd
}
