package kue

// QueuePreset contains a predefined configuration that can be applied when
// creating a new queue. The map value uses the exact SQS attribute names as
// expected by the AWS SDK CreateQueueInput Attributes field.
//
// See: https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_CreateQueue.html
//
// Only attributes that differ from the AWS default values are included.

type QueuePreset map[string]string

// Predefined queue presets that users can select from the TUI. If you add a new
// preset, make sure to also update the docs and any UI code that renders the
// list.
var Presets = map[string]QueuePreset{
	// A basic queue with very short message retention. Useful for transient work
	// queues where messages are only relevant for a brief period of time (e.g.
	// "fire-and-forget" tasks, real-time processing pipelines).
	"Short-lived": {
		"MessageRetentionPeriod": "60", // 60 seconds (minimum allowed)
	},

	// A FIFO queue optimised for very high throughput by scoping deduplication to
	// the message group instead of the entire queue. Content-based deduplication
	// is enabled so that clients do not have to supply deduplication IDs.
	"High throughput FIFO": {
		"FifoQueue":                 "true",
		"ContentBasedDeduplication": "true",
		"DeduplicationScope":        "messageGroup",
	},
}

// GetPreset returns the attribute map for the given preset name and a boolean
// indicating whether the preset exists.
func GetPreset(name string) (QueuePreset, bool) {
	preset, ok := Presets[name]
	return preset, ok
}
