package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// queueRedriveState holds the state for DLQ redrive.
type queueRedriveState struct {
	queue          kue.Queue
	destinationArn string // ARN of the original source queue to redrive messages to
	selected       int    // 0 = no, 1 = yes
	taskHandle     string
	tasks          []kue.MessageMoveTaskStatus
	inProgress     bool
	fromOverview   bool // true when triggered from queue overview
}

// running reports whether the latest redrive task is still moving messages.
func (r queueRedriveState) running() bool {
	return len(r.tasks) == 0 || r.tasks[0].Status == "RUNNING" || r.tasks[0].Status == "CANCELLING"
}

// openRedrive asks to redrive the messages of a dead-letter queue back to the queue they came from.
func (m model) openRedrive(q kue.Queue, fromOverview bool) (model, tea.Cmd) {
	sourceArn := m.findSourceQueueArn(q.Arn)
	if sourceArn == "" {
		m.error = fmt.Sprintf("queue %s is not a dead-letter queue, no queue sends its failed messages to it.", q.Name)
		return m, nil
	}
	r := &m.state.queueRedrive
	r.queue = q
	r.destinationArn = sourceArn
	r.fromOverview = fromOverview
	return m.QueueRedriveSwitchPage()
}

func (m model) QueueRedriveSwitchPage() (model, tea.Cmd) {
	m.error = ""
	r := &m.state.queueRedrive
	r.selected = 0
	r.inProgress = false
	r.taskHandle = ""
	r.tasks = nil
	return m.SwitchPage(queueRedrive), nil
}

func (m model) queueRedriveGoBack() (model, tea.Cmd) {
	if m.state.queueRedrive.fromOverview {
		return m.QueueOverviewGoBack()
	}
	if m.state.queueRedrive.inProgress {
		return m.QueueDetailsReload()
	}
	return m.queueDetailsGoBack()
}

func (m model) QueueRedriveView() string {
	if m.state.queueRedrive.inProgress {
		return m.renderRedriveProgress()
	}
	r := m.state.queueRedrive
	return dialog("redrive dead-letter queue", styles.ToneWarning,
		styles.Render(styles.S("move the messages of ", styles.ToneBody), styles.B(dialogName(r.queue.Name), styles.ToneText),
			styles.S(" back to ", styles.ToneBody), styles.B(dialogName(queueName(r.destinationArn)), styles.ToneText), styles.S("?", styles.ToneBody)),
		styles.Faint("moves about "+plural(int(atoi(r.queue.ApproximateNumberOfMessages)), "message")+", they are received again by its consumers."),
		"",
		confirmButtons(r.selected, styles.ToneWarning),
	)
}

func (m model) renderRedriveProgress() string {
	r := m.state.queueRedrive
	route := styles.Render(styles.S(r.queue.Name, styles.ToneText), styles.S(" → ", styles.ToneFaint), styles.S(queueName(r.destinationArn), styles.ToneText))

	if len(r.tasks) == 0 {
		return dialog("redrive", styles.ToneAccent, route, "", styles.Muted("waiting for status…"))
	}

	task := r.tasks[0]
	title, tone := "redrive", styles.ToneAccent
	switch task.Status {
	case "RUNNING":
		title = "redrive in progress"
	case "COMPLETED":
		title, tone = "redrive completed", styles.ToneSuccess
	case "CANCELLING":
		title, tone = "redrive cancelling", styles.ToneWarning
	case "CANCELLED":
		title, tone = "redrive cancelled", styles.ToneWarning
	case "FAILED":
		title, tone = "redrive failed", styles.ToneDanger
	}

	var ratio float64
	switch {
	case task.ApproximateNumberOfMessagesToMove > 0:
		ratio = float64(task.ApproximateNumberOfMessagesMoved) / float64(task.ApproximateNumberOfMessagesToMove)
	case task.Status == "COMPLETED":
		ratio = 1
	}
	bar := styles.Render(append(styles.Bar(ratio, 40, tone), styles.S(fmt.Sprintf(" %3.0f%%", ratio*100), styles.ToneMuted))...)
	moved := styles.Muted(fmt.Sprintf("%s of %s messages moved",
		formatCount(uint64(max(task.ApproximateNumberOfMessagesMoved, 0))),
		formatCount(uint64(max(task.ApproximateNumberOfMessagesToMove, 0)))))

	lines := []string{route, "", bar, moved}
	if task.FailureReason != "" {
		lines = append(lines, "", styles.Render(styles.S("✗ ", styles.ToneDanger), styles.S(task.FailureReason, styles.ToneDanger)))
	}
	if !r.running() {
		lines = append(lines, "", styles.Faint("press q to go back"))
	}
	return dialog(title, tone, lipgloss.JoinVertical(lipgloss.Center, lines...))
}

func (m model) QueueRedriveUpdate(msg tea.Msg) (model, tea.Cmd) {
	r := &m.state.queueRedrive
	if r.inProgress {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && key.Matches(keyMsg, m.keys.Quit, m.keys.Back) {
			return m.queueRedriveGoBack()
		}
		return m, nil
	}

	switch m.confirmKey(msg, &r.selected) {
	case confirmNo:
		return m.queueRedriveGoBack()
	case confirmYes:
		m.busy, m.loading, m.loadingMsg = true, true, "starting redrive…"
		return m, commands.StartRedrive(m.context, m.client, r.queue.Arn, r.destinationArn)
	}
	return m, nil
}

// findSourceQueueArn returns the ARN of the source queue that uses the given
// ARN as its dead-letter target, or empty string if not found.
func (m model) findSourceQueueArn(dlqArn string) string {
	for _, q := range m.state.queueOverview.queues {
		if dlqArn != "" && q.DeadLetterTargetARN == dlqArn {
			return q.Arn
		}
	}
	return ""
}
