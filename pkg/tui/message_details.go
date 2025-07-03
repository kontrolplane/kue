package tui

import (
    "fmt"
    "log"
    "os"
    "time"

    "github.com/atotto/clipboard"
    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

type queueMessageDetailsState struct {
    message kue.Message
    status  string // status line shown at bottom (e.g., saved/copied)
}

func (m model) QueueMessageDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    return m.SwitchPage(queueMessageDetails), nil
}

func (m model) QueueMessageDetailsView() string {
    if m.state.queueMessageDetails.message.MessageID == "" {
        return "No message selected"
    }

    msg := m.state.queueMessageDetails.message

    content := fmt.Sprintf("Message ID: %s\nSent: %s\nSize: %d bytes\n---\n%s",
        msg.MessageID,
        msg.SentTimestamp,
        len(msg.Body),
        msg.Body,
    )

    if m.state.queueMessageDetails.status != "" {
        content = fmt.Sprintf("%s\n\n%s", content, m.state.queueMessageDetails.status)
    }

    return content
}

func (m model) saveMessageBodyToFile() (model, tea.Cmd) {
    msg := m.state.queueMessageDetails.message
    if msg.MessageID == "" {
        m.state.queueMessageDetails.status = "No message selected"
        return m, nil
    }

    filename := fmt.Sprintf("kue-message-%s-%d.txt", msg.MessageID, time.Now().Unix())
    err := os.WriteFile(filename, []byte(msg.Body), 0644)
    if err != nil {
        log.Printf("[QueueMessageDetails] Error saving file: %v", err)
        m.state.queueMessageDetails.status = fmt.Sprintf("Failed to save message: %v", err)
    } else {
        m.state.queueMessageDetails.status = fmt.Sprintf("Saved message body to %s", filename)
    }
    return m, nil
}

func (m model) copyMessageBodyToClipboard() (model, tea.Cmd) {
    msg := m.state.queueMessageDetails.message
    if msg.MessageID == "" {
        m.state.queueMessageDetails.status = "No message selected"
        return m, nil
    }

    if err := clipboard.WriteAll(msg.Body); err != nil {
        log.Printf("[QueueMessageDetails] Error copying to clipboard: %v", err)
        m.state.queueMessageDetails.status = fmt.Sprintf("Failed to copy to clipboard: %v", err)
    } else {
        m.state.queueMessageDetails.status = "Copied message body to clipboard"
    }
    return m, nil
}

func (m model) QueueMessageDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
    var cmd tea.Cmd

    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Save):
            return m.saveMessageBodyToFile()
        case key.Matches(msg, m.keys.Copy):
            return m.copyMessageBodyToClipboard()
    case key.Matches(msg, m.keys.Quit):
        return m.QueueDetailsSwitchPage(msg)
        }
    }

    return m, cmd
}
