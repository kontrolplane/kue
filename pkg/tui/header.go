package tui

import (
	"fmt"
)

func formatHeader(projectName, programName, viewName string) string {
	return fmt.Sprintf("%s/%s • page: %s", projectName, programName, viewName)
}
