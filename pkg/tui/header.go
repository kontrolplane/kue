package tui

import (
	"fmt"
)

func formatHeader(projectName, programName, viewName string) string {
	return fmt.Sprintf("%s/%s • %s", projectName, programName, viewName)
}
