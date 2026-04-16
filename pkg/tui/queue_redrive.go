package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"

	tea "github.com/charmbracelet/bubbletea"
	kue "github.com/kontrolplane/kue/pkg/kue"
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

func (m model) QueueRedriveSwitchPage(msg tea.Msg) (model, tea.Cmd) {
	m.error = ""
	m.state.queueRedrive.selected = 0
	m.state.queueRedrive.inProgress = false
	m.state.queueRedrive.taskHandle = ""
	m.state.queueRedrive.tasks = nil
	return m.SwitchPage(queueRedrive), nil
}

func (m model) queueRedriveGoBack(msg tea.Msg) (model, tea.Cmd) {
	if m.state.queueRedrive.fromOverview {
		m.error = ""
		return m.SwitchPage(queueOverview), nil
	}
	return m.QueueDetailsGoBack(msg)
}

func (m model) QueueRedriveView() string {
	if m.state.queueRedrive.inProgress {
		return m.renderRedriveProgress()
	}
	return m.renderRedriveConfirmation()
}

func (m model) renderRedriveConfirmation() string {
	queueDisplay := styles.Bold.Render(m.state.queueRedrive.queue.Name)

	confirm := "yes"
	abort := "no"

	if m.state.queueRedrive.selected == 0 {
		abort = styles.ButtonSecondary.Render(abort)
		confirm = styles.ButtonPrimary.Render(confirm)
	} else {
		abort = styles.ButtonPrimary.Render(abort)
		confirm = styles.ButtonSecondary.Render(confirm)
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, abort, "    ", confirm)
	dialog := lipgloss.JoinVertical(lipgloss.Center,
		"warning: DLQ redrive",
		"",
		"are you sure you want to redrive messages from: "+queueDisplay+" ?",
		"",
		buttons,
	)
	return lipgloss.Place(contentWidth, contentHeight-2, lipgloss.Center, lipgloss.Center, dialog)
}

func (m model) renderRedriveProgress() string {
	labelStyle := lipgloss.NewStyle().Foreground(styles.MediumGray)
	valueStyle := lipgloss.NewStyle().Foreground(styles.TextLight)

	var lines []string

	if len(m.state.queueRedrive.tasks) == 0 {
		titleStyle := lipgloss.NewStyle().Foreground(styles.AccentColor).Bold(true)
		lines = append(lines,
			titleStyle.Render("DLQ redrive"),
			"",
			labelStyle.Render("waiting for status..."),
		)
		dialog := lipgloss.JoinVertical(lipgloss.Center, lines...)
		return lipgloss.Place(contentWidth, contentHeight-2, lipgloss.Center, lipgloss.Center, dialog)
	}

	task := m.state.queueRedrive.tasks[0]

	// Title with status-aware color
	titleColor := styles.AccentColor
	titleLabel := "DLQ redrive"
	switch task.Status {
	case "RUNNING":
		titleLabel = "DLQ redrive in progress"
	case "COMPLETED":
		titleLabel = "DLQ redrive completed"
	case "CANCELLING":
		titleLabel = "DLQ redrive cancelling"
	case "CANCELLED":
		titleLabel = "DLQ redrive cancelled"
		titleColor = styles.DangerRed
	case "FAILED":
		titleLabel = "DLQ redrive failed"
		titleColor = styles.DangerRed
	}
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(true)
	lines = append(lines, titleStyle.Render(titleLabel))
	lines = append(lines, "")

	// Queue name
	lines = append(lines, labelStyle.Render("queue: ")+valueStyle.Render(m.state.queueRedrive.queue.Name))
	lines = append(lines, "")

	// Progress bar
	barWidth := 50
	var pct float64
	if task.ApproximateNumberOfMessagesToMove > 0 {
		pct = float64(task.ApproximateNumberOfMessagesMoved) / float64(task.ApproximateNumberOfMessagesToMove)
	} else if task.Status == "COMPLETED" {
		pct = 1.0
	}

	bar := progress.New(
		progress.WithSolidFill(string(styles.AccentColor)),
		progress.WithWidth(barWidth),
	)
	bar.EmptyColor = string(styles.DarkGray)

	lines = append(lines, bar.ViewAs(pct))
	lines = append(lines, "")

	// Messages moved count
	movedText := fmt.Sprintf("%d / %d messages moved",
		task.ApproximateNumberOfMessagesMoved,
		task.ApproximateNumberOfMessagesToMove,
	)
	lines = append(lines, labelStyle.Render(movedText))

	// Failure reason
	if task.FailureReason != "" {
		lines = append(lines, "")
		failStyle := lipgloss.NewStyle().Foreground(styles.DangerRed)
		lines = append(lines, failStyle.Render("error: "+task.FailureReason))
	}

	// Navigation hint when done
	if task.Status != "RUNNING" {
		lines = append(lines, "")
		hintStyle := lipgloss.NewStyle().Foreground(styles.DarkGray)
		lines = append(lines, hintStyle.Render("press q to go back"))
	}

	content := lipgloss.JoinVertical(lipgloss.Center, lines...)
	box := lipgloss.NewStyle().
		Width(barWidth + 4).
		Align(lipgloss.Center).
		Render(content)

	return lipgloss.Place(contentWidth, contentHeight-2, lipgloss.Center, lipgloss.Center, box)
}

func (m model) switchRedriveOption() (model, tea.Cmd) {
	m.state.queueRedrive.selected = (m.state.queueRedrive.selected + 1) % 2
	return m, nil
}

func (m model) QueueRedriveUpdate(msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.state.queueRedrive.inProgress {
			if key.Matches(msg, m.keys.Quit) {
				return m.queueRedriveGoBack(msg)
			}
			return m, nil
		}
		switch {
		case key.Matches(msg, m.keys.Left):
			m, cmd = m.switchRedriveOption()
		case key.Matches(msg, m.keys.Right):
			m, cmd = m.switchRedriveOption()
		case key.Matches(msg, m.keys.View):
			if m.state.queueRedrive.selected == 0 {
				return m.queueRedriveGoBack(msg)
			}
			m.state.queueRedrive.inProgress = true
			m.loading = true
			m.loadingMsg = "Starting redrive..."
			return m, commands.StartRedrive(
				m.context, m.client,
				m.state.queueRedrive.queue.Arn,
				m.state.queueRedrive.destinationArn,
			)
		case key.Matches(msg, m.keys.Quit):
			m.state.queueRedrive.selected = 0
			return m.queueRedriveGoBack(msg)
		}
	}

	return m, cmd
}

// findSourceQueueArn returns the ARN of the source queue that uses the given
// ARN as its dead-letter target, or empty string if not found.
func (m model) findSourceQueueArn(dlqArn string) string {
	for _, q := range m.state.queueOverview.queues {
		if q.DeadLetterTargetARN == dlqArn {
			return q.Arn
		}
	}
	return ""
}
