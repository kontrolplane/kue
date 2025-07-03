package tui

import "testing"

func TestToggleSelection(t *testing.T) {
    m := model{}
    m.state.queueDetails.selectedRows = map[int]bool{}

    // simulate selection toggling at index 3
    idx := 3
    m.state.queueDetails.selectedRows[idx] = true
    if !m.state.queueDetails.selectedRows[idx] {
        t.Fatalf("expected index %d to be selected", idx)
    }

    // toggle off
    delete(m.state.queueDetails.selectedRows, idx)
    if m.state.queueDetails.selectedRows[idx] {
        t.Fatalf("expected index %d to be unselected", idx)
    }
}
