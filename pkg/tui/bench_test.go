package tui

import (
	"fmt"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/kontrolplane/kue/pkg/kue"
	"github.com/kontrolplane/kue/pkg/tui/messages"
)

// BenchmarkRender draws a frame of a queue with a page of messages, as every key and refresh does.
func BenchmarkRender(b *testing.B) {
	b.Cleanup(func() { setLayout(142, 34) })
	m, _ := update(&testing.T{}, newTestModel(), tea.WindowSizeMsg{Width: 180, Height: 50}, messages.QueuesLoadedMsg{Queues: testQueues})
	m, _ = update(&testing.T{}, m, press("j"), press("enter"))
	q := m.state.queueDetails.queue
	msgs := make([]kue.Message, 200)
	for i := range msgs {
		msgs[i] = kue.Message{
			MessageID:     fmt.Sprintf("%08d-aaaa-bbbb-cccc-dddddddddddd", i),
			Body:          fmt.Sprintf(`{"order":%d,"customer":"cus_%04d","status":"paid","items":[{"sku":"a","qty":2}],"msg":"order placed"}`, i, i),
			ReceiveCount:  "1",
			SentTimestamp: time.Now().Add(-time.Minute).Format(time.RFC3339),
		}
	}
	m, _ = update(&testing.T{}, m, messages.QueueAttributesLoadedMsg{Url: q.Url, Queue: q}, messages.MessagesLoadedMsg{Url: q.Url, Messages: msgs})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = m.render()
	}
}
