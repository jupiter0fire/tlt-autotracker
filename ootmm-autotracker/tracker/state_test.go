package tracker

import (
	"testing"
	"time"

	"ootmm-autotracker/ootmm"
)

func TestUpdateRetainsCheckNameForUncheckedDiffs(t *testing.T) {
	state := NewState()
	now := time.Unix(0, 0)
	state.now = func() time.Time { return now }
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
	if len(checks) != 0 {
		t.Fatalf("len(checks) immediately after clear = %d, want 0", len(checks))
	}

	now = now.Add(checkRemovalGracePeriod)
	_, checks, _ = state.Update(gs)
	if len(checks) != 1 {
		t.Fatalf("len(checks) after grace period = %d, want 1", len(checks))
	}
	if got := checks[0].Name; got != initialName {
		t.Fatalf("unchecked check name = %q, want %q", got, initialName)
	}
	if checks[0].Checked {
		t.Fatal("expected unchecked diff")
	}
}

func TestUpdateSuppressesTransientUncheckedDiffsWithinGracePeriod(t *testing.T) {
	state := NewState()
	now := time.Unix(0, 0)
	state.now = func() time.Time { return now }
	gs := &ootmm.GameState{}
	gs.Oot.SceneFlags[1].Chests = 1 << 2

	_, checks, _ := state.Update(gs)
	if len(checks) != 1 || !checks[0].Checked {
		t.Fatalf("initial checks = %#v, want one checked diff", checks)
	}

	gs.Oot.SceneFlags[1].Chests = 0
	_, checks, _ = state.Update(gs)
	if len(checks) != 0 {
		t.Fatalf("len(checks) after transient clear = %d, want 0", len(checks))
	}

	now = now.Add(4 * time.Second)
	gs.Oot.SceneFlags[1].Chests = 1 << 2
	_, checks, _ = state.Update(gs)
	if len(checks) != 0 {
		t.Fatalf("len(checks) after check returns within grace period = %d, want 0", len(checks))
	}

	now = now.Add(10 * time.Second)
	_, checks, _ = state.Update(gs)
	if len(checks) != 0 {
		t.Fatalf("len(checks) after grace period elapsed with check restored = %d, want 0", len(checks))
	}
}

func TestFullStateKeepsPendingChecksDuringGracePeriod(t *testing.T) {
	state := NewState()
	now := time.Unix(0, 0)
	state.now = func() time.Time { return now }
	gs := &ootmm.GameState{}
	gs.Oot.SceneFlags[1].Chests = 1 << 2

	_, _, _ = state.Update(gs)

	gs.Oot.SceneFlags[1].Chests = 0
	_, checks, _ := state.Update(gs)
	if len(checks) != 0 {
		t.Fatalf("len(checks) after starting grace period = %d, want 0", len(checks))
	}

	_, fullChecks := state.FullState(gs)
	if len(fullChecks) != 1 {
		t.Fatalf("len(fullChecks) during grace period = %d, want 1", len(fullChecks))
	}
	if !fullChecks[0].Checked {
		t.Fatal("full sync should keep pending check collected during grace period")
	}
}

func TestUpdateEmitsBothMirroredMmSkullKidCheckDiffs(t *testing.T) {
	state := NewState()
	gs := &ootmm.GameState{}
	gs.Oot.ExtraRecords[ootmm.ExtraIdxMmFlags2] = 1 << 14

	_, checks, _ := state.Update(gs)
	if len(checks) != 2 {
		t.Fatalf("len(checks) = %d, want 2", len(checks))
	}
	seen := map[string]struct{}{}
	for _, check := range checks {
		seen[check.Name] = struct{}{}
	}
	if _, ok := seen["Clock Tower Roof Skull Kid Ocarina"]; !ok {
		t.Fatal("missing Clock Tower Roof Skull Kid Ocarina diff")
	}
	if _, ok := seen["Clock Tower Roof Skull Kid Song of Time"]; !ok {
		t.Fatal("missing Clock Tower Roof Skull Kid Song of Time diff")
	}
}
