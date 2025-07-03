package util

import (
    "bytes"
    "errors"
    "os/exec"
    "runtime"
)

// CopyToClipboard attempts to copy the provided text to the system clipboard.
// It first looks for pbcopy (macOS), then xclip, and finally xsel (commonly available on Linux).
// Returns an error if no suitable clipboard program is found or if the command fails.
func CopyToClipboard(text string) error {
    // Preferred order per OS
    var candidates []string
    if runtime.GOOS == "darwin" {
        candidates = []string{"pbcopy", "xclip", "xsel"}
    } else {
        candidates = []string{"xclip", "xsel", "pbcopy"}
    }

    var prog string
    for _, c := range candidates {
        if path, err := exec.LookPath(c); err == nil {
            prog = path
            break
        }
    }

    if prog == "" {
        return errors.New("no clipboard program found (requires pbcopy, xclip, or xsel)")
    }

    cmd := exec.Command(prog)

    // For xclip/xsel we need to provide selection argument to copy to clipboard
    switch progName := basename(prog); progName {
    case "xclip":
        cmd.Args = append(cmd.Args, "-selection", "clipboard")
    case "xsel":
        cmd.Args = append(cmd.Args, "--clipboard", "--input")
    }

    cmd.Stdin = bytes.NewBufferString(text)
    return cmd.Run()
}

// basename returns the last element of a file path (similar to filepath.Base but without importing)
func basename(path string) string {
    for i := len(path) - 1; i >= 0; i-- {
        if path[i] == '/' {
            return path[i+1:]
        }
    }
    return path
}
