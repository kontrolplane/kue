package cmd

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	tui "github.com/kontrolplane/kue/pkg/tui"
)

var (
	projectName = "kontrolplane"
	programName = "kue"
)

func Execute() {
	f, err := tea.LogToFile("debug.log", "debug")
	if err != nil {
		fmt.Println("Couldn't open a file for logging:", err)
		os.Exit(1)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Println("Debug logging initialized")

	model, err := tui.NewModel(projectName, programName)
	if err != nil || model == nil {
		fmt.Println("Error creating model:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(model, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
