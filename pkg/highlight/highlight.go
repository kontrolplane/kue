package highlight

import (
    "bytes"
    "encoding/json"
    "encoding/xml"
    "fmt"

    "github.com/alecthomas/chroma/quick"
)

// DetectAndHighlight detects whether input is JSON or XML and returns the
// colorised ANSI string. If neither, it returns the raw input.
func DetectAndHighlight(body string) string {
    if body == "" {
        return "(empty message body)"
    }

    // Try JSON
    var js json.RawMessage
    if err := json.Unmarshal([]byte(body), &js); err == nil {
        buf := new(bytes.Buffer)
        if err := quick.Highlight(buf, body, "json", "terminal", "dracula"); err == nil {
            return buf.String()
        }
        // fallback return original
        return body
    }

    // Try XML
    var xm struct{}
    if err := xml.Unmarshal([]byte(body), &xm); err == nil {
        buf := new(bytes.Buffer)
        if err := quick.Highlight(buf, body, "xml", "terminal", "dracula"); err == nil {
            return buf.String()
        }
        return body
    }

    // Unknown format
    return body
}
