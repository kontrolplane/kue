package tui

import (
	"regexp"
	"strings"

	"github.com/kontrolplane/kue/pkg/tui/styles"
)

// tint splits a message body shown on one line into spans so the structure of JSON and logfmt
// recedes behind the values: keys and punctuation are faint, values keep the body ink, and the
// message, what a body is often read for, the text ink. Anything else reads as it is.
func tint(body string) cell {
	if t := strings.TrimLeft(body, " "); strings.HasPrefix(t, "{") {
		return tintJSON(body)
	}
	if c, ok := tintLogfmt(body); ok {
		return c
	}
	return cell{{Text: body, Tone: styles.ToneBody}}
}

// messageKeys are the fields loggers write the message in.
var messageKeys = []string{"msg", "message", "error", "err"}

func isMessageKey(key string) bool {
	for _, k := range messageKeys {
		if strings.EqualFold(key, k) {
			return true
		}
	}
	return false
}

// tintJSON tints JSON on one line. It does not check the JSON is valid, a broken or cut short one
// is tinted as far as it reads.
func tintJSON(line string) cell {
	var c cell
	add := func(text string, tone styles.Tone) {
		if text == "" {
			return
		}
		if n := len(c); n > 0 && c[n-1].Tone == tone {
			c[n-1].Text += text
			return
		}
		c = append(c, styles.Span{Text: text, Tone: tone})
	}
	depth := 0
	key := ""
	for i := 0; i < len(line); {
		switch ch := line[i]; {
		case ch == '"':
			j := i + 1
			for j < len(line) && line[j] != '"' {
				if line[j] == '\\' {
					j++
				}
				j++
			}
			j = min(j+1, len(line))
			k := j
			for k < len(line) && line[k] == ' ' {
				k++
			}
			switch {
			case k < len(line) && line[k] == ':':
				key = line[i+1 : max(i+1, j-1)]
				add(line[i:j], styles.ToneFaint)
			case depth == 1 && isMessageKey(key):
				add(line[i:j], styles.ToneText)
			default:
				add(line[i:j], styles.ToneBody)
			}
			i = j
		case strings.IndexByte("{}[],:", ch) >= 0:
			switch ch {
			case '{', '[':
				depth++
			case '}', ']':
				depth--
			}
			add(string(ch), styles.ToneFaint)
			i++
		default:
			j := i + 1
			for j < len(line) && strings.IndexByte(`"{}[],:`, line[j]) < 0 {
				j++
			}
			add(line[i:j], styles.ToneBody)
			i = j
		}
	}
	return c
}

// tintLogfmt tints a line of key=value pairs. It takes a line for logfmt when at least two words
// are pairs and they make up most of it, so prose with an = in it reads as it is.
func tintLogfmt(line string) (cell, bool) {
	var c cell
	words, pairs := 0, 0
	last := 0 // what is not yet in c
	for i := 0; i < len(line); {
		for i < len(line) && line[i] == ' ' {
			i++
		}
		if i >= len(line) {
			break
		}
		words++
		start := i
		for i < len(line) && line[i] != '=' && line[i] != ' ' && line[i] != '"' {
			i++
		}
		if i >= len(line) || line[i] != '=' || i == start {
			for i < len(line) && line[i] != ' ' {
				if line[i] == '"' {
					i = skipQuoted(line, i)
					continue
				}
				i++
			}
			continue
		}
		pairs++
		key := line[start:i]
		if start > last {
			c = append(c, styles.Span{Text: line[last:start], Tone: styles.ToneBody})
		}
		c = append(c, styles.Span{Text: line[start : i+1], Tone: styles.ToneFaint})
		i++ // =
		vstart := i
		if i < len(line) && line[i] == '"' {
			i = skipQuoted(line, i)
		} else {
			for i < len(line) && line[i] != ' ' {
				i++
			}
		}
		tone := styles.ToneBody
		if isMessageKey(key) {
			tone = styles.ToneText
		}
		if i > vstart {
			c = append(c, styles.Span{Text: line[vstart:i], Tone: tone})
		}
		last = i
	}
	if pairs < 2 || pairs*3 < words*2 {
		return nil, false
	}
	if last < len(line) {
		c = append(c, styles.Span{Text: line[last:], Tone: styles.ToneBody})
	}
	return c, true
}

// skipQuoted returns where the quoted string starting at i ends.
func skipQuoted(s string, i int) int {
	for j := i + 1; j < len(s); j++ {
		switch s[j] {
		case '\\':
			j++
		case '"':
			return j + 1
		}
	}
	return len(s)
}

// mark sets the byte ranges of the text of c apart, in the warm ink and bold, as a filter
// highlights its matches.
func mark(c cell, ranges [][]int) cell {
	if len(ranges) == 0 {
		return c
	}
	out := make(cell, 0, len(c)+2*len(ranges))
	pos, r := 0, 0
	for _, s := range c {
		end := pos + len(s.Text)
		for at := pos; at < end; {
			for r < len(ranges) && ranges[r][1] <= at {
				r++
			}
			cut, marked := end, false
			if r < len(ranges) {
				switch {
				case ranges[r][0] <= at:
					cut, marked = min(end, ranges[r][1]), true
				case ranges[r][0] < end:
					cut = ranges[r][0]
				}
			}
			span := styles.Span{Text: s.Text[at-pos : cut-pos], Tone: s.Tone, Bold: s.Bold}
			if marked {
				span.Tone, span.Bold = styles.ToneWarm, true
			}
			out = append(out, span)
			at = cut
		}
		pos = end
	}
	return out
}

// filterMatches is a pattern for the text a table filter matches, any case, or nil without one.
func filterMatches(filter string) *regexp.Regexp {
	if filter == "" {
		return nil
	}
	return regexp.MustCompile("(?i)" + regexp.QuoteMeta(filter))
}

// markMatches sets apart the text of c that re matches.
func markMatches(c cell, re *regexp.Regexp) cell {
	if re == nil {
		return c
	}
	var b strings.Builder
	for _, s := range c {
		b.WriteString(s.Text)
	}
	return mark(c, re.FindAllStringIndex(b.String(), -1))
}
