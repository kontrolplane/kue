package tui

import "testing"

func TestValidateQueueName(t *testing.T) {
	cases := []struct {
		name   string
		isFIFO bool
		valid  bool
	}{
		{"standard-queue", false, true},
		{"fifo-queue.fifo", true, true},
		{"fifo-queue", true, false},
		{"", false, false},
		{"this-name-is-way-too-long-because-it-exceeds-eight-zero-characters-which-is-not-allowed.fifo", true, false},
	}

	for _, c := range cases {
		err := validateQueueName(c.name, c.isFIFO)
		if (err == nil) != c.valid {
			t.Errorf("validateQueueName(%q, %v) valid expected %v got err %v", c.name, c.isFIFO, c.valid, err)
		}
	}
}
