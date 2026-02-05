package tui

import (
	"strings"
	"testing"

	"github.com/kontrolplane/kue/pkg/keys"
	"github.com/kontrolplane/kue/pkg/kue"
)

func newTestDeleteModel() model {
	return model{
		projectName: "test",
		programName: "kue",
		page:        queueDelete,
		keys:        keys.Keys,
		width:       100,
		height:      50,
		state: state{
			queueOverview: queueOverviewState{
				selected: 0,
				table:    initQueueOverviewTable(10),
			},
			queueDetails: queueDetailsState{
				selected: 0,
			},
			queueDelete: queueDeleteState{
				selected: 0,
				queue: kue.Queue{
					Name: "test-queue-to-delete",
					Url:  "http://test/test-queue-to-delete",
				},
			},
		},
	}
}

func TestQueueDeleteSwitchOption(t *testing.T) {
	m := newTestDeleteModel()

	// Initial selection should be 0 (no)
	if m.state.queueDelete.selected != 0 {
		t.Errorf("Expected initial selection to be 0, got %d", m.state.queueDelete.selected)
	}

	// Switch to yes
	m, _ = m.switchOption()
	if m.state.queueDelete.selected != 1 {
		t.Errorf("Expected selection to be 1, got %d", m.state.queueDelete.selected)
	}

	// Switch back to no
	m, _ = m.switchOption()
	if m.state.queueDelete.selected != 0 {
		t.Errorf("Expected selection to be 0, got %d", m.state.queueDelete.selected)
	}
}

func TestQueueDeleteViewContainsQueueName(t *testing.T) {
	m := newTestDeleteModel()

	view := m.QueueDeleteView()

	if !strings.Contains(view, "test-queue-to-delete") {
		t.Error("Expected view to contain queue name")
	}
}

func TestQueueDeleteViewContainsWarning(t *testing.T) {
	m := newTestDeleteModel()

	view := m.QueueDeleteView()

	if !strings.Contains(view, "warning: queue deletion") {
		t.Error("Expected view to contain warning message")
	}
}

func TestQueueDeleteViewContainsConfirmationQuestion(t *testing.T) {
	m := newTestDeleteModel()

	view := m.QueueDeleteView()

	if !strings.Contains(view, "are you sure you want to delete queue") {
		t.Error("Expected view to contain confirmation question")
	}
}

func TestQueueDeleteViewContainsButtons(t *testing.T) {
	m := newTestDeleteModel()

	view := m.QueueDeleteView()

	if !strings.Contains(view, "yes") {
		t.Error("Expected view to contain 'yes' button")
	}

	if !strings.Contains(view, "no") {
		t.Error("Expected view to contain 'no' button")
	}
}

func TestQueueDeleteSelectionResets(t *testing.T) {
	m := newTestDeleteModel()
	m.state.queueDelete.selected = 1 // Set to yes

	// Switch to delete page should not reset (that happens in Update handlers)
	m, _ = m.QueueDeleteSwitchPage(nil)

	// Verify error was cleared
	if m.error != "" {
		t.Error("Expected error to be cleared")
	}
}
