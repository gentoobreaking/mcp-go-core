package feature

import (
	"testing"
)

func TestFlagsGet(t *testing.T) {
	f := NewFlags(map[string]bool{"feature_a": true, "feature_b": false})

	if !f.Get("feature_a").Enabled {
		t.Fatal("expected feature_a to be enabled")
	}
	if f.Get("feature_b").Enabled {
		t.Fatal("expected feature_b to be disabled")
	}

	// Unknown flag returns zero-value (disabled)
	unknown := f.Get("unknown")
	if unknown.Enabled {
		t.Fatal("expected unknown flag to be disabled (zero-value)")
	}
}

func TestFlagsSet(t *testing.T) {
	f := NewFlags(map[string]bool{"feature_a": true})

	// Disable existing flag
	existed := f.Set("feature_a", false)
	if !existed {
		t.Fatal("expected Set to return true for existing flag")
	}
	if f.Get("feature_a").Enabled {
		t.Fatal("expected feature_a to be disabled after Set")
	}

	// Create new flag
	existed = f.Set("new_feature", true)
	if existed {
		t.Fatal("expected Set to return false for new flag")
	}
	if !f.Get("new_feature").Enabled {
		t.Fatal("expected new_feature to be enabled")
	}
}

func TestFlagsIsDisabled(t *testing.T) {
	f := NewFlags(map[string]bool{"enabled_feature": true, "disabled_feature": false})

	if f.IsDisabled("enabled_feature") {
		t.Fatal("expected enabled_feature to not be disabled")
	}
	if !f.IsDisabled("disabled_feature") {
		t.Fatal("expected disabled_feature to be disabled")
	}
	// Unknown flag → disabled (fail-closed)
	if !f.IsDisabled("unknown") {
		t.Fatal("expected unknown flag to be disabled (fail-closed)")
	}
}

func TestFlagsEnabledList(t *testing.T) {
	f := NewFlags(map[string]bool{"a": true, "b": false, "c": true})

	enabled := f.EnabledList()
	if len(enabled) != 2 {
		t.Fatalf("expected 2 enabled flags, got %d: %v", len(enabled), enabled)
	}

	// Snapshot should match
	snap := f.Snapshot()
	if len(snap) != 3 {
		t.Fatalf("expected 3 flags in snapshot, got %d", len(snap))
	}
	if !snap["a"].Enabled || snap["b"].Enabled || !snap["c"].Enabled {
		t.Fatal("unexpected flag states in snapshot")
	}
}
