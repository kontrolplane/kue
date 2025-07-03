package highlight

import "testing"

func TestDetectAndHighlight(t *testing.T) {
    jsonBody := `{"foo": 123}`
    xmlBody := `<note><to>You</to></note>`
    plainBody := "hello world"

    if out := DetectAndHighlight(jsonBody); out == jsonBody {
        t.Errorf("expected JSON to be highlighted, got unchanged string")
    }
    if out := DetectAndHighlight(xmlBody); out == xmlBody {
        t.Errorf("expected XML to be highlighted")
    }
    if out := DetectAndHighlight(plainBody); out != plainBody {
        t.Errorf("expected plain text to remain unchanged")
    }
}
