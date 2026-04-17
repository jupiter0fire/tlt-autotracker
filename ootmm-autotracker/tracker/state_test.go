package tracker

import (
	"testing"

	"ootmm-autotracker/ootmm"
)

func TestUpdateRetainsCheckNameForUncheckedDiffs(t *testing.T) {
	state := NewState()
	gs := &ootmm.GameState{}
	gs.Oot.SceneFlags[1].Chests = 1 << 2

	_, checks, _ := state.Update(gs)
	if len(checks) != 1 {
		t.Fatalf("len(checks) = %d, want 1", len(checks))
	}
	initialName := checks[0].Name
	if initialName == "" {
		t.Fatal("initial check name should not be empty")
	}

	gs.Oot.SceneFlags[1].Chests = 0
	_, checks, _ = state.Update(gs)
	if len(checks) != 1 {
		t.Fatalf("len(checks) after clear = %d, want 1", len(checks))
	}
	if got := checks[0].Name; got != initialName {
		t.Fatalf("unchecked check name = %q, want %q", got, initialName)
	}
	if checks[0].Checked {
		t.Fatal("expected unchecked diff")
	}
}
