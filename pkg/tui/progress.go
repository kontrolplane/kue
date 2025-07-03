package tui

import (
    tea "github.com/charmbracelet/bubbletea"
)

// progressMsg conveys progress updates from the background send operation.
// If done is true the operation has finished (successfully or not).
type progressMsg struct {
    sent int
    done bool
}

type waitProgressCmd struct {
    ch <-chan int
}

func waitProgress(ch <-chan int) tea.Cmd {
    return func() tea.Msg {
        for sent := range ch {
            // channel closed means completion
            if sent == 0 {
                continue
            }
            // We can't know total here but we send increments
            return progressMsg{sent: sent}
        }
        return progressMsg{done: true}
    }
}
