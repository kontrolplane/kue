package styles

import "testing"

func TestClean(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"orders.eu.created", "orders.eu.created"},
		{"héllo wörld ✓", "héllo wörld ✓"},
		{"a\tb", "a b"},
		{"a\nb", `a\x0ab`},
		{"a\r\nb", `a\x0d\x0ab`},
		{"\x1b[31mred\x1b[0m", `\x1b[31mred\x1b[0m`},
		{"\x1b]0;title\x07", `\x1b]0;title\x07`},
		{"del\x7f", `del\x7f`},
		{"c1\u009b2J", `c1\x9b2J`},
		{"nul\x00", `nul\x00`},
	}
	for _, tt := range tests {
		if got := Clean(tt.in); got != tt.want {
			t.Errorf("Clean(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCleanBlock(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"line one\nline two", "line one\nline two"},
		{"a\r\nb", "a\nb"},
		{"a\rb", `a\x0db`},
		{"\tindented", "    indented"},
		{"{\n\t\"k\": \"\x1b[2J\"\n}", "{\n    \"k\": \"\\x1b[2J\"\n}"},
	}
	for _, tt := range tests {
		if got := CleanBlock(tt.in); got != tt.want {
			t.Errorf("CleanBlock(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
