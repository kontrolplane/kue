package kue

import "testing"

func TestGetPreset(t *testing.T) {
    preset, ok := GetPreset("Short-lived")
    if !ok {
        t.Fatalf("expected preset to exist")
    }
    if v := preset["MessageRetentionPeriod"]; v != "60" {
        t.Errorf("unexpected MessageRetentionPeriod: got %s want 60", v)
    }

    if _, ok := GetPreset("non-existent"); ok {
        t.Errorf("expected ok=false for unknown preset")
    }
}
