package tracker

import (
	"ootmm-autotracker/ootmm"
	"time"
)

const checkRemovalGracePeriod = 5 * time.Second

// State tracks the current and previous game state, computing deltas.
type State struct {
	prevItems            map[string]int
	prevChecks           map[string]checkState
	pendingCheckRemovals map[string]pendingCheckRemoval
	prevGame             ootmm.ActiveGame
	now                  func() time.Time
}

type checkState struct {
	Name    string
	Checked bool
}

type pendingCheckRemoval struct {
	Check     checkState
	ExpiresAt time.Time
}

func NewState() *State {
	return &State{
		prevItems:            make(map[string]int),
		prevChecks:           make(map[string]checkState),
		pendingCheckRemovals: make(map[string]pendingCheckRemoval),
		now:                  time.Now,
	}
}

// ItemDiff represents a changed item.
type ItemDiff struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
}

// CheckDiff represents a changed check.
type CheckDiff struct {
	Name    string `json:"name"`
	Checked bool   `json:"checked"`
}

// Update computes what changed since the last update.
// Returns changed items, changed checks, and whether the active game changed.
func (s *State) Update(gs *ootmm.GameState) (changedItems []ItemDiff, changedChecks []CheckDiff, gameChanged bool) {
	items := ootmm.ExtractItems(gs)
	checks := ootmm.ExtractChecks(gs)
	now := s.now()

	// Items diff
	currentItems := make(map[string]int, len(items))
	for _, it := range items {
		currentItems[it.ID] = it.Qty
		prev, existed := s.prevItems[it.ID]
		delta := it.Qty - prev
		if !existed {
			delta = it.Qty
		}
		if delta != 0 {
			changedItems = append(changedItems, ItemDiff{ID: it.ID, Qty: delta})
		}
	}
	// Emit negative deltas for items that disappeared
	for id, prevQty := range s.prevItems {
		if _, exists := currentItems[id]; !exists && prevQty != 0 {
			changedItems = append(changedItems, ItemDiff{ID: id, Qty: -prevQty})
		}
	}
	s.prevItems = currentItems

	// Checks diff
	currentChecks := make(map[string]checkState, len(checks))
	for _, ch := range checks {
		current := checkState{Name: ch.Name, Checked: ch.Checked}
		currentChecks[ch.Key] = current
		delete(s.pendingCheckRemovals, ch.Key)
		prev, existed := s.prevChecks[ch.Key]
		if !existed || prev.Checked != ch.Checked {
			changedChecks = append(changedChecks, CheckDiff{Name: ch.Name, Checked: ch.Checked})
		}
		s.prevChecks[ch.Key] = current
	}
	// Delay unchecked diffs briefly so transient scene/load gaps do not flicker.
	for key, prev := range s.prevChecks {
		if _, exists := currentChecks[key]; exists {
			continue
		}
		if !prev.Checked {
			delete(s.prevChecks, key)
			delete(s.pendingCheckRemovals, key)
			continue
		}
		pending, exists := s.pendingCheckRemovals[key]
		if !exists {
			s.pendingCheckRemovals[key] = pendingCheckRemoval{
				Check:     prev,
				ExpiresAt: now.Add(checkRemovalGracePeriod),
			}
			continue
		}
		if now.Before(pending.ExpiresAt) {
			continue
		}
		changedChecks = append(changedChecks, CheckDiff{Name: prev.Name, Checked: false})
		delete(s.prevChecks, key)
		delete(s.pendingCheckRemovals, key)
	}

	// Game change
	gameChanged = gs.ActiveGame != s.prevGame
	s.prevGame = gs.ActiveGame

	return
}

// FullState returns all items and checks as if everything changed (for initial sync).
func (s *State) FullState(gs *ootmm.GameState) ([]ItemDiff, []CheckDiff) {
	items := ootmm.ExtractItems(gs)
	checks := s.effectiveChecks(gs)

	allItems := make([]ItemDiff, len(items))
	for i, it := range items {
		allItems[i] = ItemDiff{ID: it.ID, Qty: it.Qty}
	}

	allChecks := make([]CheckDiff, len(checks))
	for i, ch := range checks {
		allChecks[i] = CheckDiff{Name: ch.Name, Checked: ch.Checked}
	}

	return allItems, allChecks
}

func (s *State) effectiveChecks(gs *ootmm.GameState) []ootmm.TrackedCheck {
	checks := ootmm.ExtractChecks(gs)
	if len(s.pendingCheckRemovals) == 0 {
		return checks
	}

	now := s.now()
	seenKeys := make(map[string]struct{}, len(checks))
	seenNames := make(map[string]struct{}, len(checks))
	for _, ch := range checks {
		seenKeys[ch.Key] = struct{}{}
		seenNames[ch.Name] = struct{}{}
	}
	for key, pending := range s.pendingCheckRemovals {
		if !pending.Check.Checked || !now.Before(pending.ExpiresAt) {
			continue
		}
		if _, exists := seenKeys[key]; exists {
			continue
		}
		if _, exists := seenNames[pending.Check.Name]; exists {
			continue
		}
		checks = append(checks, ootmm.TrackedCheck{
			Key:     key,
			Name:    pending.Check.Name,
			Checked: true,
		})
		seenKeys[key] = struct{}{}
		seenNames[pending.Check.Name] = struct{}{}
	}

	return checks
}
