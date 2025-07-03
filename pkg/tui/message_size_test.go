package tui

import "testing"

func TestBytesRemaining(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected int
	}{
		{"empty", "", MaxMessageBytes},
		{"exact limit", string(make([]byte, MaxMessageBytes)), 0},
		{"one byte over", string(make([]byte, MaxMessageBytes+1)), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesRemaining(tt.body); got != tt.expected {
				t.Errorf("BytesRemaining(%q) = %d, want %d", tt.name, got, tt.expected)
			}
		})
	}
}
