package tui

import "testing"

func TestFormatJSON(t *testing.T) {
    input := `{ "foo": "bar", "number": 123 }`
    out, err := FormatJSON(input)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if out == input {
        t.Fatalf("expected formatted output to differ from input")
    }
}
