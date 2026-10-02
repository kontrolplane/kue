package tui

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/charmbracelet/x/ansi"

	"github.com/kontrolplane/kue/pkg/client"
	keys "github.com/kontrolplane/kue/pkg/keys"
	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/commands"
	"github.com/kontrolplane/kue/pkg/tui/messages"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

func NewModel(
	projectName string,
	programName string,
	sqsClient *sqs.Client,
	awsInfo client.AWSInfo,
) tea.Model {
	m := newModel(projectName, programName)
	m.client = sqsClient
	m.awsInfo = awsInfo
	return m
}

func newModel(projectName, programName string) model {
	return model{
		projectName: projectName,
		programName: programName,
		page:        queueOverview,
		context:     context.Background(),
		loading:     true,
		loadingMsg:  "loading queues…",
		spinner:     spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		keys:        keys.Keys,
		state: state{
			queueOverview: queueOverviewState{
				table:         newDataTable(queueOverviewColumns, contentWidth-4, contentHeight-tableHeaderRows),
				selectedItems: make(map[string]bool),
				filterInput:   initFilterInput("filter by name…"),
			},
			queueDetails: queueDetailsState{
				selectedItems: make(map[string]bool),
				filterInput:   initFilterInput("filter by id or body…"),
			},
		},
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(commands.LoadQueues(m.context, m.client), m.spinner.Tick)
}

// textInputActive reports whether keystrokes are currently going into a text field,
// in which case single character shortcuts must not trigger.
func (m model) textInputActive() bool {
	switch m.page {
	case queueCreate, queueMessageCreate, queueDelete:
		return true
	case queueOverview:
		return m.state.queueOverview.filtering
	case queueDetails:
		return m.state.queueDetails.filtering
	}
	return false
}

// inQueue reports whether the current page shows the queue held in the details state, so
// results loaded for that queue still belong on screen.
func (m model) inQueue() bool {
	switch m.page {
	case queueDetails, queueMessageDetails, queueMessageCreate, queueMessageDelete:
		return true
	case queuePurge:
		return !m.state.queuePurge.fromOverview
	case queueRedrive:
		return !m.state.queueRedrive.fromOverview
	}
	return false
}

// keepSpinning restarts the loading spinner when a load began after it last stopped.
func (m model) keepSpinning() (model, tea.Cmd) {
	if !m.loading || m.spinning {
		return m, nil
	}
	m.spinning = true
	return m, m.spinner.Tick
}

func (m model) scheduleRefresh() (model, tea.Cmd) {
	m.refreshGen++
	return m, commands.ScheduleRefresh(m.refreshGen)
}

// setStatus shows text in the footer for a few seconds, in the success, warning or danger tone.
func (m model) setStatus(text string, tone styles.Tone) (model, tea.Cmd) {
	m.statusGen++
	m.statusMsg = text
	m.statusTone = tone
	return m, commands.ClearStatusAfter(3*time.Second, m.statusGen)
}

// withStatus is setStatus for a handler that already has a command to return.
func (m model) withStatus(cmd tea.Cmd, text string, tone styles.Tone) (model, tea.Cmd) {
	m, status := m.setStatus(text, tone)
	return m, tea.Batch(cmd, status)
}

// finish ends a create, delete, purge or send once its result is in: it stops the spinner,
// navigates with back, and reports done in the footer, or failed with the error in a dialog.
func (m model) finish(back func(model) (model, tea.Cmd), done, failed string, err error) (model, tea.Cmd) {
	m.busy, m.loading = false, false
	m, cmd := back(m)
	if err != nil {
		m.error = fmt.Sprintf("%s: %s", failed, err)
		return m, cmd
	}
	return m.withStatus(cmd, done, styles.ToneSuccess)
}

// bulkResult describes the outcome of deleting items: what succeeded, and what to say when some
// or all of them failed.
func bulkResult(noun string, total int, deleted []string) (done, failed string) {
	label := func(names []string) string {
		if len(names) == 1 {
			return noun + " " + names[0]
		}
		return plural(len(names), noun)
	}
	done = "deleted " + label(deleted)
	if len(deleted) == 0 {
		return done, "could not delete " + plural(total, noun)
	}
	return done, fmt.Sprintf("deleted %d of %s, the others failed", len(deleted), plural(total, noun))
}

// loadError reports a failed load. Data already on screen stays there with the error in the
// footer, so a refresh failing every so often does not keep raising a dialog.
func (m model) loadError(what string, err error, onScreen bool) (model, tea.Cmd) {
	if onScreen {
		return m.setStatus(fmt.Sprintf("%s: %s", what, err), styles.ToneDanger)
	}
	m.error = fmt.Sprintf("%s: %s", what, err)
	return m, nil
}

// queueNotFound reports whether err says the queue does not exist.
func queueNotFound(err error) bool {
	var notFound *types.QueueDoesNotExist
	return errors.As(err, &notFound)
}

// queueGone leaves a queue that was deleted elsewhere.
func (m model) queueGone(name string) (model, tea.Cmd) {
	m.loading = false
	o := &m.state.queueOverview
	o.queues = deleteNames(o.queues, []string{name}, func(q kue.Queue) string { return q.Name })
	m = m.updateQueueOverviewTable()
	m, cmd := m.QueueOverviewGoBack()
	return m.withStatus(cmd, fmt.Sprintf("queue %s no longer exists", name), styles.ToneDanger)
}

// hasControlChars reports whether s holds control characters other than line breaks and tabs.
func hasControlChars(s string) bool {
	return strings.ContainsFunc(s, func(r rune) bool {
		return unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r'
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		setLayout(msg.Width, msg.Height)
		m = m.resize()

	case spinner.TickMsg:
		if !m.loading {
			m.spinning = false
			return m, nil
		}
		m.spinning = true
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.ForceQuit) {
			return m, tea.Quit
		}
		if m.busy || m.tooSmall() {
			return m, nil
		}
		if m.error != "" {
			m.error = ""
			return m, nil
		}
		if !m.textInputActive() && key.Matches(msg, m.keys.Help) {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		if !key.Matches(msg, m.keys.Back) {
			m.armed = false
		}
		if !m.textInputActive() && m.refreshes() && key.Matches(msg, m.keys.Pause) {
			return m.togglePause()
		}

	case messages.ClockTickMsg:
		m.clocking = false
		return m.keepClock()

	case messages.QueuesLoadedMsg:
		o := &m.state.queueOverview
		if m.page == queueOverview {
			m.loading = false
		}
		if msg.Err != nil {
			if m.page == queueOverview {
				m, cmd = m.loadError("could not load queues", msg.Err, o.loaded)
				cmds = append(cmds, cmd)
			}
		} else {
			prev, ok := m.currentQueue()
			o.queues = msg.Queues
			o.loaded = true
			if m.page == queueOverview {
				m.refreshedAt = time.Now()
			}
			prune(o.selectedItems, o.queues, func(q kue.Queue) string { return q.Url })
			if ok {
				o.selected = follow(m.getFilteredQueues(), o.selected, func(q kue.Queue) bool { return q.Url == prev.Url })
			}
			m = m.updateQueueOverviewTable()
		}
		if m.page == queueOverview {
			m, cmd = m.scheduleRefresh()
			cmds = append(cmds, cmd)
		}

	case messages.QueueAttributesLoadedMsg:
		d := &m.state.queueDetails
		if msg.Url != d.queue.Url || !m.inQueue() {
			break
		}
		if m.page == queueDetails {
			m.loading = false
		}
		switch {
		case queueNotFound(msg.Err) && !m.busy:
			return m.queueGone(d.queue.Name)
		case msg.Err != nil:
			m, cmd = m.loadError("could not load queue "+d.queue.Name, msg.Err, d.attributesLoaded)
			cmds = append(cmds, cmd)
		default:
			d.queue = msg.Queue
			d.attributesLoaded = true
			if m.page == queueDetails {
				m.refreshedAt = time.Now()
			}
		}

	case messages.MessagesLoadedMsg:
		d := &m.state.queueDetails
		if msg.Url != d.queue.Url || !m.inQueue() {
			break
		}
		if msg.Err != nil {
			if !queueNotFound(msg.Err) {
				m, cmd = m.loadError("could not load messages", msg.Err, d.messagesLoaded)
				cmds = append(cmds, cmd)
			}
		} else {
			prev, ok := m.currentMessage()
			d.messages = msg.Messages
			d.messagesLoaded = true
			prune(d.selectedItems, d.messages, func(msg kue.Message) string { return msg.MessageID })
			if ok {
				d.selected = follow(m.getFilteredMessages(), d.selected, func(msg kue.Message) bool { return msg.MessageID == prev.MessageID })
			}
			m = m.updateMessagesTable()
		}
		if m.page == queueDetails {
			m, cmd = m.scheduleRefresh()
			cmds = append(cmds, cmd)
		}

	case messages.QueueCreatedMsg:
		m.busy, m.loading = false, false
		if msg.Err != nil {
			return m.reopenQueueCreate(fmt.Sprintf("could not create queue: %s", msg.Err))
		}
		m, cmd = m.QueueOverviewGoBack()
		return m.withStatus(cmd, "created queue "+msg.Name, styles.ToneSuccess)

	case messages.QueuesDeletedMsg:
		o := &m.state.queueOverview
		for _, q := range o.queues {
			if slices.Contains(msg.Names, q.Name) {
				delete(o.selectedItems, q.Url)
			}
		}
		o.queues = deleteNames(o.queues, msg.Names, func(q kue.Queue) string { return q.Name })
		m = m.updateQueueOverviewTable()
		done, failed := bulkResult("queue", len(m.state.queueDelete.queues), msg.Names)
		return m.finish(model.QueueOverviewGoBack, done, failed, msg.Err)

	case messages.MessagesDeletedMsg:
		d := &m.state.queueDetails
		for _, id := range msg.MessageIDs {
			delete(d.selectedItems, id)
		}
		d.messages = deleteNames(d.messages, msg.MessageIDs, func(msg kue.Message) string { return msg.MessageID })
		m = m.updateMessagesTable()
		n, total := len(msg.MessageIDs), len(m.state.queueMessageDelete.messages)
		done := fmt.Sprintf("deleted %s from %s", plural(n, "message"), d.queue.Name)
		failed := "could not delete " + plural(total, "message")
		if n > 0 {
			failed = fmt.Sprintf("deleted %d of %s, the others failed", n, plural(total, "message"))
		}
		return m.finish(model.QueueDetailsReload, done, failed, msg.Err)

	case messages.MessageCreatedMsg:
		m.busy, m.loading = false, false
		if msg.Err != nil {
			m.error = fmt.Sprintf("could not send message: %s", msg.Err)
			return m, nil
		}
		m, cmd = m.QueueDetailsReload()
		return m.withStatus(cmd, "sent message to "+msg.Queue, styles.ToneSuccess)

	case messages.QueuePurgedMsg:
		return m.finish(model.queuePurgeFinished, "purged "+msg.Queue, "could not purge "+msg.Queue, msg.Err)

	case messages.QueueRedriveStartedMsg:
		m.busy, m.loading = false, false
		r := &m.state.queueRedrive
		if msg.Err != nil {
			r.inProgress = false
			m.error = fmt.Sprintf("could not start redrive: %s", msg.Err)
			break
		}
		r.taskHandle = msg.TaskHandle
		r.inProgress = true
		cmds = append(cmds, commands.ScheduleRedrivePoll(0, m.context, m.client, r.queue.Arn))

	case messages.QueueRedriveStatusMsg:
		r := &m.state.queueRedrive
		if m.page != queueRedrive || !r.inProgress {
			break
		}
		if msg.Err != nil {
			m, cmd = m.setStatus(fmt.Sprintf("could not poll redrive status: %s", msg.Err), styles.ToneDanger)
			cmds = append(cmds, cmd)
		} else {
			r.tasks = msg.Tasks
		}
		if msg.Err != nil || r.running() {
			cmds = append(cmds, commands.ScheduleRedrivePoll(3*time.Second, m.context, m.client, r.queue.Arn))
		}

	case messages.ClipboardCopiedMsg:
		if msg.Err != nil {
			cmds = append(cmds, tea.SetClipboard(msg.Text))
		}
		if hasControlChars(msg.Text) {
			m, cmd = m.setStatus("copied, text contains control characters", styles.ToneWarning)
		} else {
			m, cmd = m.setStatus("copied to clipboard", styles.ToneSuccess)
		}
		cmds = append(cmds, cmd)

	case messages.StatusClearMsg:
		if msg.Gen == m.statusGen {
			m.statusMsg = ""
		}

	case messages.RefreshTickMsg:
		// While paused the tick is let go, which ends the refresh loop until it resumes.
		if msg.Gen != m.refreshGen || m.paused {
			break
		}
		m, cmd = m.refreshPage()
		cmds = append(cmds, cmd)
	}

	switch m.page {
	case queueOverview:
		m, cmd = m.QueueOverviewUpdate(msg)
	case queueDetails:
		m, cmd = m.QueueDetailsUpdate(msg)
	case queueCreate:
		m, cmd = m.QueueCreateUpdate(msg)
	case queueDelete:
		m, cmd = m.QueueDeleteUpdate(msg)
	case queuePurge:
		m, cmd = m.QueuePurgeUpdate(msg)
	case queueRedrive:
		m, cmd = m.QueueRedriveUpdate(msg)
	case queueMessageDetails:
		m, cmd = m.QueueMessageDetailsUpdate(msg)
	case queueMessageDelete:
		m, cmd = m.QueueMessageDeleteUpdate(msg)
	case queueMessageCreate:
		m, cmd = m.QueueMessageCreateUpdate(msg)
	}
	cmds = append(cmds, cmd)
	m, cmd = m.keepSpinning()
	cmds = append(cmds, cmd)
	m, cmd = m.keepClock()
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// refreshes reports whether the current page refreshes on its own.
func (m model) refreshes() bool {
	switch m.page {
	case queueOverview, queueDetails:
		return true
	}
	return false
}

// refreshPage reloads what the current page shows.
func (m model) refreshPage() (model, tea.Cmd) {
	switch m.page {
	case queueOverview:
		return m, commands.LoadQueues(m.context, m.client)
	case queueDetails:
		return m, commands.LoadQueueDetails(m.context, m.client, m.state.queueDetails.queue.Url)
	}
	return m, nil
}

// togglePause stops the refresh ticks, or starts them again with a refresh.
func (m model) togglePause() (model, tea.Cmd) {
	m.paused = !m.paused
	if m.paused {
		return m, nil
	}
	return m.refreshPage()
}

// clockVisible reports whether the time since the last refresh is on screen.
func (m model) clockVisible() bool {
	return !m.refreshedAt.IsZero() && m.refreshes() && !m.loading && m.error == "" && !m.tooSmall()
}

// keepClock ticks the time since the last refresh while it is on screen, once each time its
// wording changes.
func (m model) keepClock() (model, tea.Cmd) {
	if m.clocking || !m.clockVisible() {
		return m, nil
	}
	m.clocking = true
	return m, commands.ScheduleClock(untilAgoChanges(time.Since(m.refreshedAt)))
}

// untilAgoChanges returns how long until formatAgo words an age of d differently.
func untilAgoChanges(d time.Duration) time.Duration {
	unit := time.Second
	switch {
	case d >= 24*time.Hour:
		unit = 24 * time.Hour
	case d >= time.Hour:
		unit = time.Hour
	case d >= time.Minute:
		unit = time.Minute
	}
	if d < 0 {
		return unit
	}
	return unit - d%unit
}

// formatRefreshed words the time since the last refresh.
func formatRefreshed(d time.Duration) string {
	if d < time.Second {
		return "refreshed just now"
	}
	return "refreshed " + formatAgo(time.Now().Add(-d))
}

func (m model) View() tea.View {
	v := tea.NewView(m.render())
	v.AltScreen = true
	v.WindowTitle = m.programName
	v.ForegroundColor = styles.P.Text
	if styles.Paint {
		v.BackgroundColor = styles.P.Base
	}
	return v
}

// tooSmall reports whether the terminal is smaller than the layout needs.
func (m model) tooSmall() bool {
	return m.width > 0 && (m.width < minContentWidth+chromeWidth || m.height < minContentHeight+chromeHeight)
}

// tooSmallView asks for a larger terminal, in place of a layout that would not fit.
func (m model) tooSmallView() string {
	notice := lipgloss.JoinVertical(lipgloss.Center,
		styles.Render(styles.B("terminal too small", styles.ToneWarning)),
		styles.Muted(fmt.Sprintf("%d×%d, needs %d×%d", m.width, m.height, minContentWidth+chromeWidth, minContentHeight+chromeHeight)),
		styles.Faint("ctrl+c to quit"),
	)
	return clip(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, notice), m.width, m.height)
}

func (m model) render() string {
	if m.tooSmall() {
		return m.tooSmallView()
	}

	c := m.content()
	meta, foot := m.frameMeta()
	mainView := m.renderHeader() + "\n\n" +
		frame(styles.Render(m.breadcrumb()...), meta, foot, c) + "\n" +
		m.renderFooter()

	if m.width == 0 {
		return mainView
	}
	return place(m.width, m.height, mainView)
}

// content renders what the frame holds: the page, or the error, loading or help in its place.
func (m model) content() string {
	var c string

	switch {
	case m.error != "":
		c = m.ErrorView()
	case m.loading:
		c = m.LoadingView()
	default:
		switch m.page {
		case queueOverview:
			c = m.QueueOverviewView()
		case queueDetails:
			c = m.QueueDetailsView()
		case queueCreate:
			c = m.QueueCreateView()
		case queueDelete:
			c = m.QueueDeleteView()
		case queuePurge:
			c = m.QueuePurgeView()
		case queueRedrive:
			c = m.QueueRedriveView()
		case queueMessageDetails:
			c = m.QueueMessageDetailsView()
		case queueMessageDelete:
			c = m.QueueMessageDeleteView()
		case queueMessageCreate:
			c = m.QueueMessageCreateView()
		default:
			c = errNoPageSelected
		}
	}

	if m.showHelp {
		c = m.renderHelpOverlay()
	}

	return c
}

func (m model) LoadingView() string {
	msg := m.loadingMsg
	if msg == "" {
		msg = "loading…"
	}
	return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center,
		styles.Accent(m.spinner.View())+" "+styles.Muted(truncate(styles.Clean(msg), contentWidth-4)))
}

// frameMeta returns what is set into the frame's edges: the active filter on top, the refresh
// clock and cursor position of the visible table at the bottom.
func (m model) frameMeta() (string, string) {
	var filter string
	var t *dataTable
	switch m.page {
	case queueOverview:
		filter = m.state.queueOverview.filterText
		t = &m.state.queueOverview.table
	case queueDetails:
		filter = m.state.queueDetails.filterText
		t = &m.state.queueDetails.messagesTable
	}
	var meta string
	if filter != "" {
		meta = styles.Render(styles.S("filter ", styles.ToneFaint), styles.S(filter, styles.ToneText))
	}
	if m.loading || m.error != "" {
		return meta, ""
	}
	var foot []string
	if m.clockVisible() {
		foot = append(foot, styles.Faint(formatRefreshed(time.Since(m.refreshedAt))))
	}
	if t != nil && t.position() != "" {
		foot = append(foot, styles.Faint(t.position()))
	}
	return meta, strings.Join(foot, styles.Faint(" · "))
}

func (m model) renderFooter() string {
	var status string
	if m.statusMsg != "" {
		switch m.statusTone {
		case styles.ToneDanger:
			status = styles.Render(styles.S("✗ ", styles.ToneDanger), styles.S(m.statusMsg, styles.ToneDanger))
		case styles.ToneWarning:
			status = styles.Render(styles.S("▲ ", styles.ToneWarning), styles.S(m.statusMsg, styles.ToneWarning))
		default:
			status = styles.Render(styles.S("✓ ", styles.ToneSuccess), styles.S(m.statusMsg, styles.ToneBody))
		}
	}
	if w := lipgloss.Width(status); w > frameWidth/2 {
		status = ansi.Truncate(status, frameWidth/2, "…")
	}
	width := frameWidth - 2 - lipgloss.Width(status) - 2
	left := ansi.Truncate(m.renderFooterLeft(width), width, "…")
	return " " + spread(left, status+" ", frameWidth-1)
}

// renderFooterLeft renders what the footer shows next to the status, in at most width columns.
func (m model) renderFooterLeft(width int) string {
	switch m.page {
	case queueOverview:
		o := m.state.queueOverview
		switch {
		case o.filtering:
			return m.renderFilterBar(o.filterInput.View())
		case len(o.selectedItems) > 0:
			return m.renderSelectionInfo(len(o.selectedItems), "queue")
		}
	case queueDetails:
		d := m.state.queueDetails
		switch {
		case d.filtering:
			return m.renderFilterBar(d.filterInput.View())
		case len(d.selectedItems) > 0:
			return m.renderSelectionInfo(len(d.selectedItems), "message")
		}
	}
	return m.renderShortHelp(width)
}

func (m model) renderSelectionInfo(count int, itemType string) string {
	return styles.Render(styles.S("● ", styles.ToneAccent), styles.B(plural(count, itemType)+" selected", styles.ToneText)) +
		"    " + hints([2]string{"ctrl+d", "delete"}, [2]string{"q", "clear"})
}

func (m model) renderFilterBar(inputView string) string {
	return styles.Accent("/ ") + inputView + "  " +
		hints([2]string{"enter", "apply"}, [2]string{"esc", "clear"})
}

// fitHints renders the hints that fit width. Hints are dropped from the end but for the last two,
// help and back or quit, which stay.
func fitHints(width int, pairs ...[2]string) string {
	var key strings.Builder
	fmt.Fprint(&key, width)
	for _, p := range pairs {
		key.WriteString("\x00" + p[0] + "\x00" + p[1])
	}
	return hintsMemo.get(key.String(), func() string {
		pairs = slices.Clone(pairs)
		for len(pairs) > 2 && styledWidth(hints(pairs...)) > width {
			pairs = slices.Delete(pairs, len(pairs)-3, len(pairs)-2)
		}
		return hints(pairs...)
	})
}

var hintsMemo memo

func hints(pairs ...[2]string) string {
	parts := make([]string, len(pairs))
	for i, p := range pairs {
		parts[i] = styles.Key(p[0], p[1])
	}
	return strings.Join(parts, styles.Faint("  ·  "))
}

var confirmHelp = [][2]string{{"y/n", "yes/no"}, {"←/→", "choose"}, {"enter", "confirm"}, {"esc", "cancel"}}

var shortHelp = map[page][][2]string{
	queueOverview:       {{"enter", "open"}, {"ctrl+n", "new queue"}, {"ctrl+p", "purge"}, {"ctrl+r", "redrive"}, {"ctrl+d", "delete"}, {"/", "filter"}, {"p", "pause"}, {"?", "help"}, {"q", "quit"}},
	queueDetails:        {{"enter", "open"}, {"space", "select"}, {"ctrl+n", "send"}, {"ctrl+d", "delete"}, {"ctrl+p", "purge"}, {"ctrl+r", "redrive"}, {"c", "copy arn"}, {"r", "refresh"}, {"p", "pause"}, {"/", "filter"}, {"?", "help"}, {"q", "back"}},
	queueCreate:         {{"enter", "next"}, {"shift+tab", "previous"}, {"esc", "cancel"}},
	queueDelete:         {{"enter", "delete"}, {"esc", "cancel"}},
	queuePurge:          confirmHelp,
	queueMessageDetails: {{"↑/↓", "scroll"}, {"g/G", "top/bottom"}, {"c", "copy body"}, {"ctrl+d", "delete"}, {"?", "help"}, {"q", "back"}},
	queueMessageCreate:  {{"tab", "next field"}, {"ctrl+s", "send"}, {"esc", "cancel"}},
	queueMessageDelete:  confirmHelp,
}

func (m model) renderShortHelp(width int) string {
	return fitHints(width, m.shortHelp()...)
}

func (m model) shortHelp() [][2]string {
	switch {
	case m.busy:
		// Nothing cancels an action in flight, so no key is offered.
		return nil
	case m.error != "":
		return [][2]string{{"any key", "dismiss"}}
	case m.page == queueRedrive && m.state.queueRedrive.inProgress:
		return [][2]string{{"q", "back"}}
	case m.page == queueRedrive:
		return confirmHelp
	}
	return shortHelp[m.page]
}

// renderHelpOverlay draws the key reference in a card, as roomy as the content area allows.
func (m model) renderHelpOverlay() string {
	var overlay string
	for _, fit := range []struct{ padY, padX, gap int }{{1, 4, 6}, {1, 2, 3}, {0, 2, 3}} {
		overlay = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.P.RuleBold).
			Padding(fit.padY, fit.padX).
			Render(m.renderHelpContent(fit.gap))
		if lipgloss.Width(overlay) <= contentWidth && lipgloss.Height(overlay) <= contentHeight {
			break
		}
	}
	return lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, overlay)
}

// renderHelpContent lays the help out in two columns.
func (m model) renderHelpContent(gap int) string {
	title := func(s string) string {
		return styles.Fg(styles.ToneAccent).Bold(true).MarginBottom(1).Render(s)
	}
	keyStyle := styles.Fg(styles.ToneText).Bold(true).Width(11)
	row := func(key, desc string) string {
		return keyStyle.Render(key) + styles.Muted(desc)
	}

	navigation := lipgloss.JoinVertical(lipgloss.Left,
		title("navigation"),
		row("↑/k ↓/j", "move"),
		row("g / G", "first / last"),
		row("pgup/pgdn", "page"),
		row("enter", "open"),
		row("tab/⇧tab", "next / previous field"),
		row("y / n", "answer a dialog"),
		row("q", "back, quit"),
		row("esc", "back, clear filter"),
		row("ctrl+c", "quit"),
	)

	actions := lipgloss.JoinVertical(lipgloss.Left,
		title("actions"),
		row("space", "select"),
		row("c", "copy body or arn"),
		row("ctrl+n", "new queue, send message"),
		row("ctrl+d", "delete"),
		row("ctrl+p", "purge"),
		row("ctrl+r", "redrive dead-letter queue"),
		row("r / p", "refresh / pause"),
		row("/", "filter"),
		row("?", "help"),
	)

	columns := lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().MarginRight(gap).Render(navigation), actions)
	return lipgloss.JoinVertical(lipgloss.Center, columns, "", styles.Faint("press any key to close"))
}
