package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	ProjectName = "kontrolplane"
	ProgramName = "kue"
)

// Init is the first function that will be called. It returns an optional
// initial command. To not perform an initial command return nil.
func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle(ProgramName)
}

func Execute() {

	// Context is a context that can be used to cancel the program
	ctx := context.Background()

	// Get the list of queues
	queues, err := ListQueues(ctx)

	// Initialize the model
	m := model{
		cursor:   0,
		keys:     keys,
		help:     help.New(),
		selected: make(map[int]struct{}),

		queues: queues,
	}

	// Setup debug logging to file
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}
	defer f.Close()

	// Run the program
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

}
