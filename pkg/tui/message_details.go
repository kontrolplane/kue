package tui

import (
    "fmt"

    "github.com/charmbracelet/bubbles/key"
    "github.com/charmbracelet/bubbles/viewport"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/kontrolplane/kue/pkg/highlight"
)

type queueMessageDetailsState struct {
    viewport viewport.Model
    message  string
}

func (m model) initMessageViewport() {
    width := m.width - 4  // some padding
    height := m.height - 10 // header/footer space
    if width <= 0 {
        width = 80
    }
    if height <= 0 {
        height = 20
    }
    content := highlight.DetectAndHighlight(m.state.queueMessageDetails.message)
    m.state.queueMessageDetails.viewport = viewport.New(width, height)
    m.state.queueMessageDetails.viewport.SetContent(content)
    m.state.queueMessageDetails.viewport.WrapText = false
}

func (m model) QueueMessageDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    selected := m.state.queueDetails.selected
    if selected < len(m.state.queueDetails.messages) {
        m.state.queueMessageDetails.message = m.state.queueDetails.messages[selected].Body
    }
    m.initMessageViewport()
    return m.SwitchPage(queueMessageDetails), nil
}

func (m model) QueueMessageDetailsView() string {
    vpView := m.state.queueMessageDetails.viewport.View()
    hint := "←/→/↑/↓ to scroll · q to back"
    return fmt.Sprintf("%s\n\n%s", vpView, hint)
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.initMessageViewport()
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit):
            // go back to details list
            return m.QueueDetailsSwitchPage(msg)
        default:
            m.state.queueMessageDetails.viewport, cmd = m.state.queueMessageDetails.viewport.Update(msg)
        }
    default:
        m.state.queueMessageDetails.viewport, cmd = m.state.queueMessageDetails.viewport.Update(msg)
    }

    return m, cmd
}
