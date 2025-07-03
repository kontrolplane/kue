package tui

import (
    "context"
    "fmt"
    "log"

    "strings"

    "github.com/charmbracelet/bubbles/key"
    tea "github.com/charmbracelet/bubbletea"
    kue "github.com/kontrolplane/kue/pkg/kue"
)

// queueDetailsState stores the queue currently selected for displaying
// its attributes in the details panel.
type queueDetailsState struct {
    queue kue.Queue
}

// QueueDetailsSwitchPage is invoked from the queue overview when the user
// presses the "view" (enter) key. It refreshes the attributes for the
// selected queue (so the information is up-to-date) and switches to the
// details page.
func (m model) QueueDetailsSwitchPage(msg tea.Msg) (model, tea.Cmd) {
    log.Println("[QueueDetailsSwitchPage]")

    q := m.state.queueDetails.queue
    if q.Url == "" {
        // Nothing selected; just return to overview
        return m, nil
    }

    // Refresh attributes each time we open the panel to ensure fresh data.
    refreshed, err := kue.FetchQueueAttributes(m.client, context.Background(), q.Url)
    if err != nil {
        m.error = fmt.Sprintf("Error fetching queue attributes: %v", err)
    } else {
        m.state.queueDetails.queue = refreshed
    }

    return m.SwitchPage(queueDetails), nil
}

// QueueDetailsUpdate handles key events while the details panel is active. The
// only interaction needed for this read-only view is to quit back to the
// overview.
func (m model) QueueDetailsUpdate(msg tea.Msg) (model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch {
        case key.Matches(msg, m.keys.Quit):
            return m.QueueOverviewSwitchPage(msg)
        }
    }
    return m, nil
}

// QueueDetailsView renders the queue attribute information in a simple
// human-readable block.
func (m model) QueueDetailsView() string {
    q := m.state.queueDetails.queue
    if q.Url == "" {
        return "No queue selected"
    }

    view := fmt.Sprintf(
        "Queue Information\n\n"+
            "Name:  %s\n"+
            "URL:   %s\n"+
            "ARN:   %s\n"+
            "Region:%s\n"+
            "Created:%s\n"+
            "#Msgs: %s\n",
        q.Name,
        q.Url,
        q.Arn,
        extractRegionFromArn(q.Arn),
        q.CreatedTimestamp,
        q.ApproximateNumberOfMessages,
    )

    return view
}

// extractRegionFromArn pulls the AWS region from the ARN string. If the ARN is
// malformed or empty, an empty string is returned.
func extractRegionFromArn(arn string) string {
    // arn:partition:service:region:account-id:resource
    parts := strings.Split(arn, ":")
    if len(parts) >= 4 {
        return parts[3]
    }
    return ""
}
