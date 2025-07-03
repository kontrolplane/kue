package audit

import (
    "encoding/json"
    "fmt"
    "os"
    "os/user"
    "path/filepath"
    "sync"
    "time"
)

// logEntry represents a single audit log line written in JSON format.
// Fields are exported to ensure they are marshalled properly.
type logEntry struct {
    Queue     string `json:"queue"`
    MessageID string `json:"messageId"`
    Timestamp string `json:"timestamp"`
    User      string `json:"user"`
}

var (
    // mu guards concurrent writes that may initialise the log directory for the first time.
    mu sync.Mutex
)

// AppendDeletion writes a single JSON-formatted line that represents a message deletion event to the
// audit log. The line is appended to a file located at "$HOME/.kue/audit.log". The function is best
// effort – it returns an error only if the write fails, but it should never block the deletion
// operation itself.
func AppendDeletion(queue, messageID string) error {
    userStr := currentUser()

    entry := logEntry{
        Queue:     queue,
        MessageID: messageID,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        User:      userStr,
    }

    payload, err := json.Marshal(entry)
    if err != nil {
        return fmt.Errorf("unable to marshal audit entry: %w", err)
    }

    // Ensure directory exists – protect with mutex so that concurrent goroutines don't race when
    // creating the directory for the first time.
    mu.Lock()
    defer mu.Unlock()

    homeDir, err := os.UserHomeDir()
    if err != nil {
        return fmt.Errorf("cannot determine home directory: %w", err)
    }

    dirPath := filepath.Join(homeDir, ".kue")
    if err := os.MkdirAll(dirPath, 0o755); err != nil {
        return fmt.Errorf("unable to create audit directory: %w", err)
    }

    filePath := filepath.Join(dirPath, "audit.log")
    f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
    if err != nil {
        return fmt.Errorf("unable to open audit log: %w", err)
    }
    defer f.Close()

    if _, err := f.Write(append(payload, '\n')); err != nil {
        return fmt.Errorf("unable to write audit entry: %w", err)
    }

    // Flush to disk to make sure we do not lose events if the application crashes.
    if err := f.Sync(); err != nil {
        return fmt.Errorf("unable to sync audit log: %w", err)
    }

    return nil
}

// currentUser attempts to determine the current IAM or OS user that triggered the deletion. The IAM
// user can require an AWS call, which would slow down the fast-path considerably, so for now we rely
// on the operating-system user. This can be revisited later if stricter attribution is required.
func currentUser() string {
    if u, err := user.Current(); err == nil {
        return u.Username
    }
    // Fallback to environment variable if os/user lookup fails (e.g., in cross-compiled binaries).
    if env := os.Getenv("USER"); env != "" {
        return env
    }
    return "unknown"
}
