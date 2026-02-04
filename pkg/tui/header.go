package tui

import (
	"fmt"

	"github.com/kontrolplane/kue/pkg/client"
)

func formatHeader(projectName, programName, viewName string, awsInfo client.AWSInfo) string {
	return fmt.Sprintf("%s/%s • %s • profile: %s | region: %s",
		projectName, programName, viewName, awsInfo.Profile, awsInfo.Region)
}
