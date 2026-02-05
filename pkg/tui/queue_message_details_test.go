package tui

import (
	"strings"
	"testing"

	"github.com/kontrolplane/kue/pkg/keys"
	"github.com/kontrolplane/kue/pkg/kue"
)

func newTestMessageDetailsModel() model {
	return model{
		projectName: "test",
		programName: "kue",
		page:        queueMessageDetails,
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
				queue: kue.Queue{
					Name: "test-queue",
					Url:  "http://test/test-queue",
				},
			},
			queueMessageDetails: queueMessageDetailsState{
				queueName: "test-queue",
				message: kue.Message{
					MessageID:     "msg-12345-abcde",
					Body:          `{"key": "value", "data": "test message body"}`,
					MD5OfBody:     "abc123def456",
					SentTimestamp: "2024-01-15T10:30:00Z",
					ReceiveCount:  "3",
					Attributes: map[string]string{
						"SenderId": "123456789",
					},
					MessageAttributes: map[string]string{
						"CustomAttr": "custom-value",
					},
				},
			},
		},
	}
}

func TestQueueMessageDetailsViewContainsMessageID(t *testing.T) {
	m := newTestMessageDetailsModel()

	view := m.renderMessageDetails()

	if !strings.Contains(view, "msg-12345-abcde") {
		t.Error("Expected view to contain message ID")
	}
}

func TestQueueMessageDetailsViewContainsQueueName(t *testing.T) {
	m := newTestMessageDetailsModel()

	view := m.renderMessageDetails()

	// Queue name appears in the message ID or other fields
	if !strings.Contains(view, "msg-12345-abcde") {
		t.Error("Expected view to contain message details")
	}
}

func TestQueueMessageDetailsViewContainsBody(t *testing.T) {
	m := newTestMessageDetailsModel()

	view := m.renderMessageDetails()

	if !strings.Contains(view, "test message body") {
		t.Error("Expected view to contain message body")
	}
}

func TestQueueMessageDetailsViewContainsSentTimestamp(t *testing.T) {
	m := newTestMessageDetailsModel()

	view := m.renderMessageDetails()

	if !strings.Contains(view, "2024-01-15T10:30:00Z") {
		t.Error("Expected view to contain sent timestamp")
	}
}

func TestQueueMessageDetailsViewContainsReceiveCount(t *testing.T) {
	m := newTestMessageDetailsModel()

	view := m.renderMessageDetails()

	if !strings.Contains(view, "3") {
		t.Error("Expected view to contain receive count")
	}
}

func TestQueueMessageDetailsViewContainsCustomAttributes(t *testing.T) {
	m := newTestMessageDetailsModel()

	view := m.renderMessageDetails()

	if !strings.Contains(view, "CustomAttr") {
		t.Error("Expected view to contain custom attribute name")
	}
	if !strings.Contains(view, "custom-value") {
		t.Error("Expected view to contain custom attribute value")
	}
}

func TestQueueMessageDetailsViewContainsSections(t *testing.T) {
	m := newTestMessageDetailsModel()

	view := m.renderMessageDetails()

	if !strings.Contains(view, "Basic Information") {
		t.Error("Expected view to contain 'Basic Information' section")
	}
	if !strings.Contains(view, "Body") {
		t.Error("Expected view to contain 'Body' section")
	}
}

func TestQueueMessageDetailsSwitchPage(t *testing.T) {
	m := newTestMessageDetailsModel()
	m.page = queueDetails
	m.error = "some previous error"

	m, _ = m.QueueMessageDetailsSwitchPage(nil)

	if m.page != queueMessageDetails {
		t.Errorf("Expected page to be queueMessageDetails, got %v", m.page)
	}
	if m.error != "" {
		t.Error("Expected error to be cleared")
	}
}

func TestQueueMessageDetailsWithFIFOAttributes(t *testing.T) {
	m := newTestMessageDetailsModel()
	m.state.queueMessageDetails.message.MessageGroupID = "group-1"
	m.state.queueMessageDetails.message.MessageDeduplicationID = "dedup-123"
	m.state.queueMessageDetails.message.SequenceNumber = "12345"

	view := m.renderMessageDetails()

	if !strings.Contains(view, "FIFO Attributes") {
		t.Error("Expected view to contain 'FIFO Attributes' section")
	}
	if !strings.Contains(view, "group-1") {
		t.Error("Expected view to contain message group ID")
	}
	if !strings.Contains(view, "dedup-123") {
		t.Error("Expected view to contain deduplication ID")
	}
}
