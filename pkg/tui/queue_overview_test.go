package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kontrolplane/kue/pkg/keys"
	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/messages"
)

func newTestModel() model {
	return model{
		projectName: "test",
		programName: "kue",
		page:        queueOverview,
		keys:        keys.Keys,
		help:        help.New(),
		width:       100,
		height:      50,
		state: state{
			queueOverview: queueOverviewState{
				selected: 0,
				queues:   nil,
				table:    initQueueOverviewTable(10),
			},
			queueDetails: queueDetailsState{
				selected: 0,
			},
			queueDelete: queueDeleteState{
				selected: 0,
			},
		},
	}
}

func TestQueueOverviewNavigation(t *testing.T) {
	m := newTestModel()

	// Add some test queues
	m.state.queueOverview.queues = []kue.Queue{
		{Name: "queue-1", Url: "http://test/queue-1"},
		{Name: "queue-2", Url: "http://test/queue-2"},
		{Name: "queue-3", Url: "http://test/queue-3"},
	}

	// Test moving down
	m, _ = m.nextQueue()
	if m.state.queueOverview.selected != 1 {
		t.Errorf("Expected selected to be 1, got %d", m.state.queueOverview.selected)
	}

	m, _ = m.nextQueue()
	if m.state.queueOverview.selected != 2 {
		t.Errorf("Expected selected to be 2, got %d", m.state.queueOverview.selected)
	}

	// Test boundary - should not go beyond last item
	m, _ = m.nextQueue()
	if m.state.queueOverview.selected != 2 {
		t.Errorf("Expected selected to stay at 2, got %d", m.state.queueOverview.selected)
	}

	// Test moving up
	m, _ = m.previousQueue()
	if m.state.queueOverview.selected != 1 {
		t.Errorf("Expected selected to be 1, got %d", m.state.queueOverview.selected)
	}

	m, _ = m.previousQueue()
	if m.state.queueOverview.selected != 0 {
		t.Errorf("Expected selected to be 0, got %d", m.state.queueOverview.selected)
	}

	// Test boundary - should not go below 0
	m, _ = m.previousQueue()
	if m.state.queueOverview.selected != 0 {
		t.Errorf("Expected selected to stay at 0, got %d", m.state.queueOverview.selected)
	}
}

func TestQueuesLoadedMsgHandler(t *testing.T) {
	m := newTestModel()

	testQueues := []kue.Queue{
		{Name: "queue-1", Url: "http://test/queue-1", ApproximateNumberOfMessages: "5"},
		{Name: "queue-2", Url: "http://test/queue-2", ApproximateNumberOfMessages: "10"},
	}

	// Simulate receiving QueuesLoadedMsg
	msg := messages.QueuesLoadedMsg{Queues: testQueues, Err: nil}
	newModel, _ := m.Update(msg)

	updatedModel := newModel.(model)

	if len(updatedModel.state.queueOverview.queues) != 2 {
		t.Errorf("Expected 2 queues, got %d", len(updatedModel.state.queueOverview.queues))
	}

	if updatedModel.loading {
		t.Error("Expected loading to be false after QueuesLoadedMsg")
	}

	if updatedModel.error != "" {
		t.Errorf("Expected no error, got: %s", updatedModel.error)
	}
}

func TestQueuesLoadedMsgWithError(t *testing.T) {
	m := newTestModel()

	// Simulate receiving QueuesLoadedMsg with error
	msg := messages.QueuesLoadedMsg{Queues: nil, Err: tea.ErrProgramKilled}
	newModel, _ := m.Update(msg)

	updatedModel := newModel.(model)

	if updatedModel.error == "" {
		t.Error("Expected error to be set")
	}

	if updatedModel.loading {
		t.Error("Expected loading to be false after error")
	}
}

func TestNoQueuesFound(t *testing.T) {
	m := newTestModel()
	m.state.queueOverview.queues = nil

	if !m.NoQueuesFound() {
		t.Error("Expected NoQueuesFound to return true when queues is nil")
	}

	m.state.queueOverview.queues = []kue.Queue{}
	if !m.NoQueuesFound() {
		t.Error("Expected NoQueuesFound to return true when queues is empty")
	}

	m.state.queueOverview.queues = []kue.Queue{{Name: "test"}}
	if m.NoQueuesFound() {
		t.Error("Expected NoQueuesFound to return false when queues exist")
	}
}

func TestQueueOverviewViewWithQueues(t *testing.T) {
	m := newTestModel()
	m.state.queueOverview.queues = []kue.Queue{
		{Name: "test-queue", Url: "http://test/test-queue"},
	}

	view := m.QueueOverviewView()

	// View should not contain "No queues found" message
	if view == "No queues found. Press Ctrl+N to create a new queue." {
		t.Error("Expected table view, not empty queue message")
	}
}

func TestQueueOverviewViewWithoutQueues(t *testing.T) {
	m := newTestModel()
	m.state.queueOverview.queues = nil

	view := m.QueueOverviewView()

	// View should contain the empty message centered in the table area
	if !strings.Contains(view, "No queues found. Press Ctrl+N to create a new queue.") {
		t.Errorf("Expected view to contain empty queue message, got '%s'", view)
	}
}
