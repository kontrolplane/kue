package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/messages"
	"github.com/kontrolplane/kue/pkg/tui/styles"
)

func newTestModel() model {
	m := newModel("test", "kue")
	m.width = 160
	m.height = 50
	m.loading = false
	return m
}

func press(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "space":
		return tea.KeyPressMsg{Code: tea.KeySpace, Text: " "}
	}
	if ctrl, ok := strings.CutPrefix(s, "ctrl+"); ok {
		return tea.KeyPressMsg{Code: rune(ctrl[0]), Mod: tea.ModCtrl}
	}
	return tea.KeyPressMsg{Code: rune(s[0]), Text: s}
}

func update(t *testing.T, m model, msgs ...tea.Msg) (model, tea.Cmd) {
	t.Helper()
	var cmd tea.Cmd
	for _, msg := range msgs {
		var next tea.Model
		next, cmd = m.Update(msg)
		m = next.(model)
	}
	return m, cmd
}

func typeText(t *testing.T, m model, text string) model {
	t.Helper()
	for _, r := range text {
		m, _ = update(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func isQuit(cmd tea.Cmd) bool {
	if cmd == nil {
		return false
	}
	msg := cmd()
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if isQuit(c) {
				return true
			}
		}
		return false
	}
	_, ok := msg.(tea.QuitMsg)
	return ok
}

const arnPrefix = "arn:aws:sqs:us-east-1:000000000000:"

var testQueues = []kue.Queue{
	{Name: "emails", Url: "http://test/emails", Arn: arnPrefix + "emails", ApproximateNumberOfMessages: "3", VisibilityTimeout: "30", MessageRetentionPeriod: "345600"},
	{Name: "orders", Url: "http://test/orders", Arn: arnPrefix + "orders", ApproximateNumberOfMessages: "240", DeadLetterTargetARN: arnPrefix + "orders-dlq", RedrivePolicy: `{"deadLetterTargetArn":"` + arnPrefix + `orders-dlq","maxReceiveCount":5}`},
	{Name: "orders-dlq", Url: "http://test/orders-dlq", Arn: arnPrefix + "orders-dlq", ApproximateNumberOfMessages: "2"},
	{Name: "payments.fifo", Url: "http://test/payments.fifo", Arn: arnPrefix + "payments.fifo", FifoQueue: "true", ContentBasedDeduplication: "true"},
}

var testMessages = []kue.Message{
	{MessageID: "11111111-aaaa", Body: `{"order":1,"ok":true}`, ReceiptHandle: "r1", ReceiveCount: "1", SentTimestamp: time.Now().Add(-time.Hour).Format(time.RFC3339), MessageAttributes: map[string]string{"source": "web"}},
	{MessageID: "22222222-bbbb", Body: "plain text body", ReceiptHandle: "r2", ReceiveCount: "4"},
}

func loadedModel(t *testing.T) model {
	t.Helper()
	m, _ := update(t, newTestModel(), messages.QueuesLoadedMsg{Queues: testQueues})
	return m
}

// detailsModel opens the details of orders with its messages loaded.
func detailsModel(t *testing.T) model {
	t.Helper()
	m := loadedModel(t)
	m, _ = update(t, m, press("j"), press("enter"))
	q := m.state.queueDetails.queue
	m, _ = update(t, m,
		messages.QueueAttributesLoadedMsg{Url: q.Url, Queue: q},
		messages.MessagesLoadedMsg{Url: q.Url, Messages: testMessages},
	)
	return m
}

func TestQueuesLoaded(t *testing.T) {
	m := newTestModel()
	m.loading = true
	m, cmd := update(t, m, messages.QueuesLoadedMsg{Queues: testQueues})
	if m.loading {
		t.Error("expected loading to stop once the queues are in")
	}
	if cmd == nil {
		t.Error("expected the next refresh to be scheduled")
	}
	if got := len(m.state.queueOverview.table.Rows()); got != len(testQueues) {
		t.Errorf("expected %d rows, got %d", len(testQueues), got)
	}
	view := m.render()
	for _, want := range []string{"emails", "orders-dlq", "payments.fifo", "dlq", "000000000000"} {
		if !strings.Contains(ansi.Strip(view), want) {
			t.Errorf("expected the overview to show %q", want)
		}
	}
}

func TestQueuesLoadErrorKeepsDataOnScreen(t *testing.T) {
	m, _ := update(t, newTestModel(), messages.QueuesLoadedMsg{Err: errors.New("boom")})
	if m.error == "" {
		t.Error("expected a first failed load to raise an error dialog")
	}

	m = loadedModel(t)
	m, _ = update(t, m, messages.QueuesLoadedMsg{Err: errors.New("boom")})
	if m.error != "" {
		t.Error("expected a failed refresh to leave the loaded queues on screen")
	}
	if !strings.Contains(m.statusMsg, "boom") || m.statusTone != styles.ToneDanger {
		t.Errorf("expected the failure in the footer, got %q", m.statusMsg)
	}
	if len(m.state.queueOverview.queues) != len(testQueues) {
		t.Error("expected the queues to be kept")
	}
}

func TestEmptyOverview(t *testing.T) {
	m, _ := update(t, newTestModel(), messages.QueuesLoadedMsg{})
	if !strings.Contains(ansi.Strip(m.QueueOverviewView()), "no queues yet") {
		t.Error("expected the empty overview to explain how to create a queue")
	}
}

func TestOverviewNavigation(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("j"), press("j"))
	if got := m.state.queueOverview.selected; got != 2 {
		t.Errorf("expected the cursor on row 2, got %d", got)
	}
	m, _ = update(t, m, press("j"), press("j"), press("j"))
	if got := m.state.queueOverview.selected; got != len(testQueues)-1 {
		t.Errorf("expected the cursor to stop on the last row, got %d", got)
	}
	m, _ = update(t, m, press("g"))
	if got := m.state.queueOverview.selected; got != 0 {
		t.Errorf("expected g to go to the first row, got %d", got)
	}
}

func TestOverviewQuitAndBack(t *testing.T) {
	m := loadedModel(t)
	if _, cmd := update(t, m, press("esc")); isQuit(cmd) {
		t.Error("expected esc not to quit")
	}
	if _, cmd := update(t, m, press("q")); !isQuit(cmd) {
		t.Error("expected q to quit on the overview")
	}
	if _, cmd := update(t, m, press("ctrl+c")); !isQuit(cmd) {
		t.Error("expected ctrl+c to quit")
	}

	m, _ = update(t, m, press("space"))
	m, cmd := update(t, m, press("q"))
	if isQuit(cmd) || len(m.state.queueOverview.selectedItems) != 0 {
		t.Error("expected q to clear the selection before quitting")
	}
}

func TestOverviewFilter(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("/"))
	m = typeText(t, m, "order")
	if got := len(m.state.queueOverview.table.Rows()); got != 2 {
		t.Fatalf("expected the filter to leave 2 queues, got %d", got)
	}
	m, _ = update(t, m, press("enter"))
	if m.state.queueOverview.filtering || m.state.queueOverview.filterText != "order" {
		t.Error("expected enter to apply the filter")
	}
	if meta, _ := m.frameMeta(); !strings.Contains(ansi.Strip(meta), "order") {
		t.Error("expected the frame to show the filter")
	}
	m, _ = update(t, m, press("esc"))
	if m.state.queueOverview.filterText != "" || len(m.state.queueOverview.table.Rows()) != len(testQueues) {
		t.Error("expected esc to clear the filter")
	}
}

func TestSelectionFollowsQueueAcrossRefresh(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("j"), press("space"))
	reordered := []kue.Queue{testQueues[3], testQueues[1], testQueues[0]}
	m, _ = update(t, m, messages.QueuesLoadedMsg{Queues: reordered})
	names, _ := m.getSelectedQueues()
	if len(names) != 1 || names[0] != "orders" {
		t.Errorf("expected orders to stay selected, got %v", names)
	}
	if q, _ := m.currentQueue(); q.Name != "orders" {
		t.Errorf("expected the cursor to stay on orders, got %s", q.Name)
	}
}

func TestDeleteQueueAsksForTheName(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("ctrl+d"))
	if m.page != queueDelete {
		t.Fatalf("expected the delete dialog, got page %d", m.page)
	}
	m = typeText(t, m, "email")
	m, _ = update(t, m, press("enter"))
	if m.busy {
		t.Fatal("expected a partial name not to delete the queue")
	}
	m = typeText(t, m, "s")
	m, cmd := update(t, m, press("enter"))
	if !m.busy || cmd == nil {
		t.Fatal("expected the typed name to delete the queue")
	}

	m, _ = update(t, m, messages.QueuesDeletedMsg{Names: []string{"emails"}})
	if m.busy || m.page != queueOverview {
		t.Error("expected to return to the overview once deleted")
	}
	if !strings.Contains(m.statusMsg, "deleted queue emails") {
		t.Errorf("expected a status, got %q", m.statusMsg)
	}
	if len(m.state.queueOverview.queues) != len(testQueues)-1 {
		t.Error("expected the deleted queue to leave the list")
	}
}

func TestDeleteSelectionHiddenByFilter(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("space"), press("/"))
	m = typeText(t, m, "orders")
	m, _ = update(t, m, press("enter"), press("ctrl+d"))
	if m.page == queueDelete {
		t.Error("expected no dialog when the whole selection is hidden")
	}
	if !strings.Contains(m.statusMsg, "hidden by the filter") {
		t.Errorf("expected a warning, got %q", m.statusMsg)
	}
}

func TestPurgeAsksTwiceForLargeQueues(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("j"), press("ctrl+p"))
	if m.page != queuePurge {
		t.Fatal("expected the purge dialog")
	}
	m, _ = update(t, m, press("y"))
	if !m.state.queuePurge.secondPrompt || m.busy {
		t.Fatal("expected a second prompt for a queue with 240 messages")
	}
	m, cmd := update(t, m, press("y"))
	if !m.busy || cmd == nil {
		t.Fatal("expected the second yes to purge")
	}
	m, _ = update(t, m, messages.QueuePurgedMsg{Queue: "orders"})
	if m.page != queueOverview || !strings.Contains(m.statusMsg, "purged orders") {
		t.Errorf("expected to return to the overview with a status, got page %d and %q", m.page, m.statusMsg)
	}
}

func TestPurgeSmallQueueAsksOnce(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("ctrl+p"), press("y"))
	if !m.busy {
		t.Error("expected a queue with 3 messages to purge after one yes")
	}
}

func TestRedriveNeedsADeadLetterQueue(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("ctrl+r"))
	if m.page == queueRedrive || !strings.Contains(m.error, "not a dead-letter queue") {
		t.Errorf("expected an error for a queue that is not a dead-letter queue, got %q", m.error)
	}

	m = loadedModel(t)
	m, _ = update(t, m, press("j"), press("j"), press("ctrl+r"))
	if m.page != queueRedrive {
		t.Fatal("expected the redrive dialog for orders-dlq")
	}
	if m.state.queueRedrive.destinationArn != arnPrefix+"orders" {
		t.Errorf("expected to redrive to orders, got %s", m.state.queueRedrive.destinationArn)
	}
	m, _ = update(t, m, press("y"), messages.QueueRedriveStartedMsg{TaskHandle: "t"})
	m, _ = update(t, m, messages.QueueRedriveStatusMsg{Tasks: []kue.MessageMoveTaskStatus{{Status: "RUNNING", ApproximateNumberOfMessagesMoved: 1, ApproximateNumberOfMessagesToMove: 2}}})
	if view := ansi.Strip(m.QueueRedriveView()); !strings.Contains(view, "1 of 2 messages moved") {
		t.Errorf("expected the progress, got:\n%s", view)
	}
}

func TestQueueDetails(t *testing.T) {
	m := detailsModel(t)
	if m.page != queueDetails || m.loading {
		t.Fatalf("expected the loaded details, page %d loading %v", m.page, m.loading)
	}
	view := ansi.Strip(m.render())
	for _, want := range []string{"queues › orders", "orders-dlq", "after 5 receives", "11111111-aaaa", "plain text body"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected the details to show %q", want)
		}
	}

	m, cmd := update(t, m, press("q"))
	if m.page != queueOverview || cmd == nil {
		t.Error("expected q to return to the overview and refresh it")
	}
}

func TestQueueDetailsIgnoresOtherQueues(t *testing.T) {
	m := detailsModel(t)
	m, _ = update(t, m, messages.MessagesLoadedMsg{Url: "http://test/emails"})
	if len(m.state.queueDetails.messages) != len(testMessages) {
		t.Error("expected messages of another queue to be dropped")
	}
}

func TestMessageDetails(t *testing.T) {
	m := detailsModel(t)
	m, _ = update(t, m, press("enter"))
	if m.page != queueMessageDetails {
		t.Fatal("expected the message details")
	}
	view := ansi.Strip(m.render())
	for _, want := range []string{"11111111-aaaa", "orders", `"order": 1`, "source", "web", "json"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected the message details to show %q", want)
		}
	}
	m, _ = update(t, m, press("esc"))
	if m.page != queueDetails {
		t.Error("expected esc to return to the details")
	}
}

func TestMessageDetailsFifo(t *testing.T) {
	m := detailsModel(t)
	m.state.queueMessageDetails.message = kue.Message{MessageID: "x", MessageGroupID: "group-a", SequenceNumber: "42"}
	m, _ = m.QueueMessageDetailsSwitchPage()
	view := ansi.Strip(m.QueueMessageDetailsView())
	for _, want := range []string{"fifo", "group-a", "42"} {
		if !strings.Contains(view, want) {
			t.Errorf("expected the fifo attributes to show %q", want)
		}
	}
}

func TestDeleteMessages(t *testing.T) {
	m := detailsModel(t)
	m, _ = update(t, m, press("space"), press("j"), press("space"), press("ctrl+d"))
	if m.page != queueMessageDelete || len(m.state.queueMessageDelete.messages) != 2 {
		t.Fatalf("expected to delete the 2 selected messages, page %d", m.page)
	}
	m, cmd := update(t, m, press("y"))
	if !m.busy || cmd == nil {
		t.Fatal("expected y to delete")
	}
	m, _ = update(t, m, messages.MessagesDeletedMsg{MessageIDs: []string{"11111111-aaaa"}, Err: errors.New("denied")})
	if m.page != queueDetails || !strings.Contains(m.error, "deleted 1 of 2 messages") {
		t.Errorf("expected a partial failure, got %q", m.error)
	}
	if len(m.state.queueDetails.messages) != 1 {
		t.Error("expected the deleted message to leave the list")
	}
}

func TestSendMessage(t *testing.T) {
	m := detailsModel(t)
	m, _ = update(t, m, press("ctrl+n"))
	if m.page != queueMessageCreate {
		t.Fatal("expected the send form")
	}
	m, _ = update(t, m, press("ctrl+s"))
	if m.busy || m.error == "" {
		t.Error("expected an empty body to be refused")
	}
	m, _ = update(t, m, press("x"))

	m = typeText(t, m, `{"a":1}`)
	if !strings.Contains(ansi.Strip(m.render()), "json ✓") {
		t.Error("expected the body to be recognised as json")
	}
	m, _ = update(t, m, press("esc"))
	if m.page != queueMessageCreate || m.statusMsg != discardPrompt {
		t.Fatal("expected esc to ask before discarding the body")
	}
	m, _ = update(t, m, press("ctrl+s"))
	if !m.busy {
		t.Fatal("expected ctrl+s to send")
	}
	m, _ = update(t, m, messages.MessageCreatedMsg{Queue: "orders"})
	if m.page != queueDetails || !strings.Contains(m.statusMsg, "sent message to orders") {
		t.Errorf("expected to return to the details, got page %d and %q", m.page, m.statusMsg)
	}
}

func TestCreateQueueDiscard(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("ctrl+n"))
	if m.page != queueCreate {
		t.Fatal("expected the create form")
	}
	m = typeText(t, m, "new")
	m, _ = update(t, m, press("esc"))
	if m.page != queueCreate {
		t.Fatal("expected the first esc to ask before discarding")
	}
	m, _ = update(t, m, press("esc"))
	if m.page != queueOverview {
		t.Error("expected the second esc to discard the form")
	}
}

func TestCreateQueueFailureKeepsInput(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("ctrl+n"))
	m = typeText(t, m, "new")
	m.busy = true
	m, _ = update(t, m, messages.QueueCreatedMsg{Name: "new", Err: errors.New("exists")})
	if m.page != queueCreate || m.state.queueCreate.input.name != "new" || !strings.Contains(m.error, "exists") {
		t.Errorf("expected the form back with its input and the error, got page %d", m.page)
	}
}

func TestBuildQueueConfig(t *testing.T) {
	cfg, err := buildQueueConfig(&queueCreateInput{name: " jobs ", queueType: "fifo", visibilityTimeout: "60", contentBasedDeduplication: true, deduplicationScope: "queue"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Name != "jobs" || !cfg.IsFifo || cfg.VisibilityTimeout != 60 || !cfg.ContentBasedDeduplication {
		t.Errorf("unexpected config %+v", cfg)
	}
	if _, err := buildQueueConfig(&queueCreateInput{name: "no spaces"}); err == nil {
		t.Error("expected an invalid name to be refused")
	}
}

func TestHelpAndErrorDismiss(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("?"))
	if !m.showHelp || !strings.Contains(ansi.Strip(m.render()), "navigation") {
		t.Fatal("expected ? to show the help")
	}
	m, _ = update(t, m, press("j"))
	if m.showHelp || m.state.queueOverview.selected != 0 {
		t.Error("expected any key to close the help without acting")
	}

	m.error = "boom"
	m, _ = update(t, m, press("j"))
	if m.error != "" || m.state.queueOverview.selected != 0 {
		t.Error("expected any key to dismiss the error without acting")
	}
}

func TestPauseStopsRefresh(t *testing.T) {
	m := loadedModel(t)
	m, _ = update(t, m, press("p"))
	if !m.paused || !strings.Contains(ansi.Strip(m.renderHeader()), "paused") {
		t.Fatal("expected p to pause")
	}
	_, cmd := update(t, m, messages.RefreshTickMsg{Gen: m.refreshGen})
	for _, msg := range collect(cmd) {
		if _, ok := msg.(messages.QueuesLoadedMsg); ok {
			t.Error("expected no refresh while paused")
		}
	}
}

// collect runs cmd and returns the messages it produces, skipping ticks.
func collect(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	select {
	case msg := <-done:
		if batch, ok := msg.(tea.BatchMsg); ok {
			var out []tea.Msg
			for _, c := range batch {
				out = append(out, collect(c)...)
			}
			return out
		}
		return []tea.Msg{msg}
	case <-time.After(50 * time.Millisecond):
		return nil
	}
}

func TestLayoutFits(t *testing.T) {
	t.Cleanup(func() { setLayout(142, 34) })
	for _, size := range [][2]int{{102, 25}, {140, 40}, {220, 60}} {
		w, h := size[0], size[1]
		pages := map[string]model{}
		open := func(name string, keys ...string) {
			m := detailsModel(t)
			m, _ = update(t, m, tea.WindowSizeMsg{Width: w, Height: h})
			for _, k := range keys {
				m, _ = update(t, m, press(k))
			}
			pages[name] = m
		}
		open("details")
		open("message details", "enter")
		open("send", "ctrl+n")
		open("message delete", "ctrl+d")
		open("purge", "ctrl+p")
		open("help", "?")
		open("overview", "q")
		open("create", "q", "ctrl+n")
		open("delete", "q", "ctrl+d")

		for name, m := range pages {
			lines := strings.Split(m.render(), "\n")
			if len(lines) > h {
				t.Errorf("%dx%d %s: %d lines, more than the terminal's %d", w, h, name, len(lines), h)
			}
			for i, line := range lines {
				if lw := ansi.StringWidth(line); lw > w {
					t.Errorf("%dx%d %s: line %d is %d wide", w, h, name, i, lw)
				}
			}
		}
	}
}

func TestTooSmall(t *testing.T) {
	m := loadedModel(t)
	t.Cleanup(func() { setLayout(142, 34) })
	m, _ = update(t, m, tea.WindowSizeMsg{Width: 60, Height: 10})
	if !strings.Contains(ansi.Strip(m.render()), "terminal too small") {
		t.Error("expected a notice for a small terminal")
	}
}

func TestFormatHelpers(t *testing.T) {
	if got := formatSeconds("345600"); got != "4d" {
		t.Errorf("formatSeconds: got %q", got)
	}
	if got := formatSeconds("90"); got != "1m30s" {
		t.Errorf("formatSeconds: got %q", got)
	}
	if got := accountID(arnPrefix + "orders"); got != "000000000000" {
		t.Errorf("accountID: got %q", got)
	}
	if parseTime("2024-01-02 03:04:05").IsZero() || parseTime(time.Now().Format(time.RFC3339)).IsZero() {
		t.Error("parseTime: expected both timestamp formats to parse")
	}
}
