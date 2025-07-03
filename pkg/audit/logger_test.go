package audit

import (
    "bufio"
    "encoding/json"
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestAppendDeletion(t *testing.T) {
    // Setup a temp directory to avoid writing to real HOME.
    tmpDir := t.TempDir()
    t.Setenv("HOME", tmpDir)

    queue := "test-queue"
    msgID := "abc-123"

    if err := AppendDeletion(queue, msgID); err != nil {
        t.Fatalf("append deletion failed: %v", err)
    }

    // Verify file exists and contains exactly one line with correct JSON.
    filePath := filepath.Join(tmpDir, ".kue", "audit.log")
    f, err := os.Open(filePath)
    if err != nil {
        t.Fatalf("cannot open audit file: %v", err)
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    if !scanner.Scan() {
        t.Fatalf("audit log is empty")
    }
    line := scanner.Text()

    if scanner.Scan() {
        t.Fatalf("audit log contains more than one line; expected one")
    }

    var e logEntry
    if err := json.NewDecoder(strings.NewReader(line)).Decode(&e); err != nil {
        t.Fatalf("cannot decode json: %v", err)
    }

    if e.Queue != queue {
        t.Errorf("queue mismatch: got %s, want %s", e.Queue, queue)
    }
    if e.MessageID != msgID {
        t.Errorf("message id mismatch: got %s, want %s", e.MessageID, msgID)
    }
    if e.User == "" {
        t.Errorf("user should be populated")
    }
    if e.Timestamp == "" {
        t.Errorf("timestamp should be populated")
    }
}
