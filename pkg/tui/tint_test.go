package tui

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// spans renders a cell as text|tone pairs, with bold spans starred, to compare in tests.
func spans(c cell) string {
	parts := make([]string, len(c))
	for i, s := range c {
		star := ""
		if s.Bold {
			star = "*"
		}
		parts[i] = fmt.Sprintf("%s%q:%d", star, s.Text, s.Tone)
	}
	return strings.Join(parts, " ")
}

func joined(c cell) string {
	var b strings.Builder
	for _, s := range c {
		b.WriteString(s.Text)
	}
	return b.String()
}

func TestTint(t *testing.T) {
	F, B, T := styles.ToneFaint, styles.ToneBody, styles.ToneText
	tests := []struct {
		line string
		want cell
	}{
		{`{"level":"info","msg":"hi \"there\"","n":3,"o":{"msg":"x"}}`, cell{
			{Text: `{"level":`, Tone: F}, {Text: `"info"`, Tone: B}, {Text: `,"msg":`, Tone: F},
			{Text: `"hi \"there\""`, Tone: T}, {Text: `,"n":`, Tone: F}, {Text: `3`, Tone: B},
			{Text: `,"o":{"msg":`, Tone: F}, {Text: `"x"`, Tone: B}, {Text: `}}`, Tone: F},
		}},
		{`level=info msg="request done" path=/api status=200`, cell{
			{Text: `level=`, Tone: F}, {Text: `info`, Tone: B}, {Text: ` `, Tone: B}, {Text: `msg=`, Tone: F},
			{Text: `"request done"`, Tone: T}, {Text: ` `, Tone: B}, {Text: `path=`, Tone: F}, {Text: `/api`, Tone: B},
			{Text: ` `, Tone: B}, {Text: `status=`, Tone: F}, {Text: `200`, Tone: B},
		}},
		{`the answer is x=42 according to the docs`, cell{{Text: `the answer is x=42 according to the docs`, Tone: B}}},
		{`{"broken":`, cell{{Text: `{"broken":`, Tone: F}}},
	}
	for _, tt := range tests {
		got := tint(tt.line)
		if spans(got) != spans(tt.want) {
			t.Errorf("tint(%s)\n got %s\nwant %s", tt.line, spans(got), spans(tt.want))
		}
		if joined(got) != tt.line {
			t.Errorf("tint(%s) changed the text to %s", tt.line, joined(got))
		}
	}
}

func TestMark(t *testing.T) {
	c := tint(`a=1 msg=timeout b=2`)
	marked := mark(c, [][]int{{2, 3}, {8, 12}, {13, 15}})
	if joined(marked) != `a=1 msg=timeout b=2` {
		t.Fatalf("mark changed the text: %s", spans(marked))
	}
	var hot []string
	for _, s := range marked {
		if s.Bold && s.Tone == styles.ToneWarm {
			hot = append(hot, s.Text)
		}
	}
	if got := strings.Join(hot, "|"); got != "1|time|ut" {
		t.Errorf("marked %q, in %s", got, spans(marked))
	}
}

func TestBodyPreviewTintsAndMarks(t *testing.T) {
	body := "{\n  \"order\": 1,\n  \"msg\": \"Order Placed\"\n}"
	c := bodyPreview(body, filterMatches("order"))
	if got := joined(c); got != `{ "order": 1, "msg": "Order Placed" }` {
		t.Fatalf("preview = %q", got)
	}
	var hot []string
	for _, s := range c {
		if s.Bold && s.Tone == styles.ToneWarm {
			hot = append(hot, s.Text)
		}
	}
	if got := strings.Join(hot, "|"); got != "order|Order" {
		t.Errorf("expected both cases of the filter set apart, got %q in %s", got, spans(c))
	}
	if c := bodyPreview(body, nil); slices.ContainsFunc(c, func(s styles.Span) bool { return s.Bold }) {
		t.Errorf("without a filter nothing is set apart: %s", spans(c))
	}
}
